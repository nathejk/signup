package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/nathejk/shared-go/messages"
	"github.com/nathejk/shared-go/types"
	jsonapi "nathejk.dk/cmd/api/app"
	"nathejk.dk/internal/data"
	"nathejk.dk/internal/validator"
	"nathejk.dk/pkg/streaminterface"
)

/*
	func (app *application) homeHandler(w http.ResponseWriter, r *http.Request) {
		config := map[string]any{
			"timeCountdown": app.config.countdown.time,
			"videos":        app.config.countdown.videos,
		}
		err := app.WriteJSON(w, http.StatusOK, jsonapi.Envelope{"config": config}, nil)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
		}
	}
*/

type Department struct {
	Label       string
	Slug        string
	Dapartments []Department
}

func (app *application) showSignupHandler(w http.ResponseWriter, r *http.Request) {
	id := types.TeamID(app.ReadNamedParam(r, "id"))
	if id == "" {
		log.Print("Not IDea")
		app.NotFoundResponse(w, r)
		return
	}
	team, err := app.models.Signup.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			log.Printf("Not Found %q", id)
			app.NotFoundResponse(w, r)
		default:
			app.ServerErrorResponse(w, r, err)
		}
		return
	}
	err = app.WriteJSON(w, http.StatusOK, jsonapi.Envelope{"signup": team}, nil)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}

func (app *application) signupPincodeHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TeamID  types.TeamID `json:"teamId"`
		Pincode string       `json:"pincode"`
	}
	if err := app.ReadJSON(w, r, &input); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	team, err := app.models.Signup.GetByID(input.TeamID)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	if team.Pincode != input.Pincode {
		app.InvalidCredentialsResponse(w, r)
	}
	page := fmt.Sprintf("/%s/%s", team.TeamType, input.TeamID)
	err = app.WriteJSON(w, http.StatusCreated, jsonapi.Envelope{"team": map[string]string{"teamPage": page}}, nil)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}

func (app *application) sendMobilepaySmsHandler(w http.ResponseWriter, r *http.Request) {
	teamID := types.TeamID(app.ReadNamedParam(r, "id"))
	if teamID == "" {
		log.Println("1")
		app.NotFoundResponse(w, r)
		return
	}
	var input struct {
		Amount int               `json:"amount"`
		Phone  types.PhoneNumber `json:"phone"`
	}
	if err := app.ReadJSON(w, r, &input); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	team, err := app.models.Signup.GetByID(teamID)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	mobilepay := 330811
	if team.TeamType == types.TeamTypePatrulje {
		mobilepay = 204414
	}
	text := fmt.Sprintf("https://www.mobilepay.dk/erhverv/betalingslink/betalingslink-svar?phone=%d&amount=%d&comment=%s&lock=1", mobilepay, input.Amount, teamID)
	err = app.sms.Send(input.Phone.Normalize(), text)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	err = app.WriteJSON(w, http.StatusCreated, jsonapi.Envelope{"ok": true}, nil)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}
func (app *application) confirmSignupHandler(w http.ResponseWriter, r *http.Request) {
	id := app.ReadNamedParam(r, "id")
	if id == "" {
		app.NotFoundResponse(w, r)
		return
	}
	teamID, err := app.models.Signup.ConfirmBySecret(id)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
		return
	}
	team, err := app.models.Signup.GetByID(teamID)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	err = app.sms.Send(team.PhonePending.Normalize(), "Din aktiveringskode til Nathejktilmeldingen er: "+team.Pincode)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/indskrivning/%s", teamID), http.StatusSeeOther)
}

func (app *application) commandCreatePerson(person *data.Personnel) {
	if person.Pincode == "" {
		pin := fmt.Sprintf("%v", rand.Float64())
		pin = pin[len(pin)-4:]
		person.Pincode = pin
	}
	if person.UserID == "" {
		person.UserID = types.UserID("user-" + uuid.New().String())
	}

	err := app.sms.Send(person.Phone.Normalize(), "Din pinkode til Nathejktilmeldingen er: "+person.Pincode)
	if err != nil {
		//app.BadRequestResponse(w, r, err)
		return
	}
	msg := app.stan.MessageFunc()(streaminterface.SubjectFromStr("nathejk:personnel.updated"))
	msg.SetBody(&messages.NathejkPersonnelCreated{
		UserID:  person.UserID,
		Phone:   types.PhoneNumber(person.Phone.Normalize()),
		Pincode: person.Pincode,
	})
	msg.SetMeta(&messages.Metadata{Producer: "deltag-api"})
	if err := app.stan.Publish(msg); err != nil {
		app.logger.PrintError(err, nil)
	}
}
func (app *application) startHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PhonePending types.PhoneNumber `json:"phonePending"`
	}
	if err := app.ReadJSON(w, r, &input); err != nil {
		log.Print(err)
		app.BadRequestResponse(w, r, err)
		return
	}
	person, err := app.models.Personnel.GetByPhone(input.PhonePending)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			person = &data.Personnel{Phone: input.PhonePending}
		default:
			app.ServerErrorResponse(w, r, err)
			return
		}
	}
	app.Background(func() {
		app.commandCreatePerson(person)
	})
	/*
		msg := &messages.NathejkTeamSignedUp{
			TeamID: input.TeamID,
			Name:   input.Name,
			Phone:  input.PhonePending,
			Email:  input.EmailPending,
		}*/

	err = app.WriteJSON(w, http.StatusCreated, jsonapi.Envelope{"status": "ok"}, nil)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}

func (app *application) verifyHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Phone   types.PhoneNumber `json:"phonePending"`
		Pincode string            `json:"pincode"`
	}
	if err := app.ReadJSON(w, r, &input); err != nil {
		log.Print(err)
		app.BadRequestResponse(w, r, err)
		return
	}
	person, err := app.models.Personnel.GetByPhone(input.Phone)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
	err = app.WriteJSON(w, http.StatusCreated, jsonapi.Envelope{"userId": person.UserID}, nil)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}
func (app *application) showPersonHandler(w http.ResponseWriter, r *http.Request) {
	id := types.UserID(app.ReadNamedParam(r, "id"))
	if id == "" {
		log.Print("Not IDea")
		app.NotFoundResponse(w, r)
		return
	}
	person, err := app.models.Personnel.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			log.Printf("Not Found %q", id)
			app.NotFoundResponse(w, r)
		default:
			app.ServerErrorResponse(w, r, err)
		}
		return
	}
	err = app.WriteJSON(w, http.StatusOK, jsonapi.Envelope{"person": person}, nil)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}
func (app *application) signupHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TeamID       types.TeamID       `json:"teamId"`
		TeamType     types.TeamType     `json:"type"`
		Name         string             `json:"name"`
		EmailPending types.EmailAddress `json:"emailPending"`
		PhonePending types.PhoneNumber  `json:"phonePending"`
	}
	/*
		var response struct {
			TeamID       types.TeamID      `json:"teamId"`
			TeamType     types.TeamType    `json:"type"`
			Name         string            `json:"name"`
			EmailPending types.Email       `json:"emailPending"`
			PhonePending types.PhoneNumber `json:"phonePending"`
		}
	*/
	if err := app.ReadJSON(w, r, &input); err != nil {
		log.Print(err)
		app.BadRequestResponse(w, r, err)
		return
	}

	msg := &messages.NathejkTeamSignedUp{
		TeamID: input.TeamID,
		Name:   input.Name,
		Phone:  input.PhonePending,
		Email:  input.EmailPending,
	}
	/*
	   v := validator.New()

	   	if product.Validate(v); !v.Valid() {
	   		app.FailedValidationResponse(w, r, v.Errors)
	   		return
	   	}
	*/
	if err := app.commands.Team.Signup(input.TeamType, msg); err != nil {
		spew.Dump(input)
		app.ServerErrorResponse(w, r, err)
		return
	}

	team, _ := app.models.Signup.GetByID(msg.TeamID)
	if team == nil {
		team = &data.Signup{TeamID: msg.TeamID, TeamType: input.TeamType}
	}
	team.Name = input.Name
	team.PhonePending = input.PhonePending
	team.EmailPending = input.EmailPending

	app.Background(func() {
		data := map[string]any{
			"team":   team,
			"secret": uuid.New().String(),
		}

		err := app.mailer.Send(string(input.EmailPending), "verify_email.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
		msg := app.stan.MessageFunc()(streaminterface.SubjectFromStr(fmt.Sprintf("NATHEJK:%s.%s.%s.mail.%s.sent", "2024", input.TeamType, team.TeamID, types.PingTypeSignup)))
		msg.SetBody(&messages.NathejkMailSent{
			PingType:  types.PingTypeSignup,
			TeamID:    team.TeamID,
			Recipient: types.EmailAddress(input.EmailPending),
			Subject:   "Bekræft e-mailadresse",
		})
		msg.SetMeta(&messages.Metadata{Producer: "deltag-api", Phase: data["secret"].(string)})
		if err := app.stan.Publish(msg); err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	err := app.WriteJSON(w, http.StatusCreated, jsonapi.Envelope{"team": team}, nil)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}

func (app *application) updatePersonHandler(w http.ResponseWriter, r *http.Request) {
	id := types.UserID(app.ReadNamedParam(r, "id"))
	if id == "" {
		log.Print("Not IDea")
		app.NotFoundResponse(w, r)
		return
	}
	person, err := app.models.Personnel.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.NotFoundResponse(w, r)
		default:
			app.ServerErrorResponse(w, r, err)
		}
		return
	}
	// If the request contains a X-Expected-Version header, verify that the
	// version in the database matches the expected version specified in the header.
	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(person.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.EditConflictResponse(w, r)
			return
		}
	}

	var input struct {
		Name       *string             `json:"name"`
		Email      *types.EmailAddress `json:"email"`
		Phone      *types.PhoneNumber  `json:"phone"`
		Korps      *types.CorpsSlug    `json:"corps"`
		MedlemNr   *string             `json:"medlemNr"`
		Department *string             `json:"department"`
		Pincode    *string             `json:"pincode"`
		Diet       *string             `json:"diet"`
	}
	if err = app.ReadJSON(w, r, &input); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	if input.Name != nil {
		person.Name = *input.Name
	}
	if input.Email != nil {
		person.Email = *input.Email
	}
	if input.Phone != nil {
		person.Phone = *input.Phone
	}
	if input.Korps != nil {
		person.Korps = *input.Korps
	}
	if input.MedlemNr != nil {
		person.MedlemNr = *input.MedlemNr
	}
	if input.Department != nil {
		person.Department = *input.Department
	}
	if input.Pincode != nil {
		person.Pincode = *input.Pincode
	}
	if input.Diet != nil {
		person.Diet = *input.Diet
	}
	v := validator.New()
	if person.Validate(v); !v.Valid() {
		app.FailedValidationResponse(w, r, v.Errors)
		return
	}
	msg := app.stan.MessageFunc()(streaminterface.SubjectFromStr("nathejk:personnel.updated"))
	msg.SetBody(&messages.NathejkPersonnelUpdated{
		UserID:     person.UserID,
		Name:       person.Name,
		Email:      person.Email,
		Phone:      types.PhoneNumber(person.Phone.Normalize()),
		Corps:      person.Korps,
		MedlemNr:   person.MedlemNr,
		Department: person.Department,
		Pincode:    person.Pincode,
		Diet:       person.Diet,
	})
	msg.SetMeta(&messages.Metadata{Producer: "deltag-api"})
	if err := app.stan.Publish(msg); err != nil {
		app.logger.PrintError(err, nil)
	}

	/*
		err = app.models.Products.Update(product)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrEditConflict):
				app.EditConflictResponse(w, r)
			default:
				app.ServerErrorResponse(w, r, err)
			}
			return
		}*/
	err = app.WriteJSON(w, http.StatusOK, jsonapi.Envelope{"person": person}, nil)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}

/*
func (app *application) deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.NotFoundResponse(w, r)
		return
	}
	err = app.models.Products.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.NotFoundResponse(w, r)
		default:
			app.ServerErrorResponse(w, r, err)
		}
		return
	}

	err = app.WriteJSON(w, http.StatusOK, jsonapi.Envelope{"message": "product successfully deleted"}, nil)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}

func (app *application) listProductsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}
	v := validator.New()

	qs := r.URL.Query()

	input.Title = app.ReadString(qs, "title", "")
	input.Genres = app.ReadCSV(qs, "genres", []string{})

	input.Filters.Page = app.ReadInt(qs, "page", 1, v)
	input.Filters.PageSize = app.ReadInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.ReadString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	if input.Filters.Validate(v); !v.Valid() {
		app.FailedValidationResponse(w, r, v.Errors)
		return
	}

	products, metadata, err := app.models.Products.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
		return
	}

	err = app.WriteJSON(w, http.StatusOK, jsonapi.Envelope{"metadata": metadata, "products": products}, nil)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}
*/
