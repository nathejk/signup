<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import Navigation from '@/components/Navigation.vue'
import Stepper from 'primevue/stepper';
import StepperPanel from 'primevue/stepperpanel';
import FloatLabel from 'primevue/floatlabel';
import InputIcon from 'primevue/inputicon';
import IconField from 'primevue/iconfield';
import InputOtp from 'primevue/inputotp';

const props = defineProps({
    memberId: {type: String, required: false},
    teamType: {type: String, required: false, default:'spejder'},
})

const router = useRouter()
const person = ref({ })
const contact = ref({type: props.teamType })
const step = ref(0);
const config = ref({
    corps:[
        { slug: "dds", label: "Det Danske Spejderkorps" },
        { slug: "kfum", label: "KFUM-Spejderne" },
        { slug: "kfuk", label: "De grønne pigespejdere" },
        { slug: "dbs", label: "Danske Baptisters Spejderkorps" },
        { slug: "dgs", label: "De Gule Spejdere" },
        { slug: "dss", label: "Dansk Spejderkorps Sydslesvig" },
        { slug: "fdf", label: "FDF / FPF" },
        { slug: "andet", label: "Andet" },
    ],
    diets:[
        { slug: "all", label: "Ingen særlige ønsker" },
        { slug: "vegetarian", label: "Vegetar" },
    ],
    areas:[
        { slug: "bandit", label: "Banditter" },
        { slug: "guide", label: "Guide" },
        { slug: "post", label: "Postmandskab" },
        { slug: "logistik", label: "Logistik" },
        { slug: "teknik", label: "Teknisk Tjeneste" },
        { slug: "pr", label: "PR og kommunikation" },
        { slug: "kok", label: "Køkkenhold" },
        { slug: "andet", label: "Andet" },
    ],
});

onMounted(async () => {
    if (!props.memberId) {
        return
    }
    try {
        const response = await fetch("/api/person/" + props.memberId);
        if (!response.ok) {
            throw new Error("HTTP status " + response.status);
        }
        const data = await response.json();
        person.value = data.person
    } catch (error) {
        console.log("mounted load failed", error);
    }
})

const complete = computed(() => !!person.value.name )
const emailValidated = computed(() => !!contact.value.email && contact.value.email == contact.value.emailPending)

const save = async () => {
    const headers = {
        "Content-Type": "application/json",
    }
    try {
        const body = JSON.stringify(person.value)
        const response = await fetch("/api/person/" + person.value.id, { method: 'PATCH', body: body, headers: headers });
        if (!response.ok) {
            throw new Error("HTTP status " + response.status);
        }
        //const data = await response.json();
        //contact.value = data.team
        //router.replace({ name: 'indskrivning', params: { id: data.teamId } })
        router.push({ path: '/tak' })

        //const vendor = data.content
    } catch (error) {
        console.log("team signup failed", error);
    }
}
const showpin = ref(false);
const pincode = ref();
const validatePincode = async () => {
    const headers = {
        "Content-Type": "application/json",
    }
    console.log('pincode', pincode)
    try {
        const body = JSON.stringify({
            phone: contact.value.phone,
            pincode: contact.value.pincode,
        })
        const response = await fetch("/api/verify", { method: 'POST', body: body, headers: headers });
        if (!response.ok) {
            throw new Error("HTTP status " + response.status);
        }
        const data = await response.json();
        router.replace({ path: '/indskrivning/'+ data.person.userId  })
        //router.replace({ path: data.team.teamPage })
    } catch (error) {
        console.log("team signup failed", error);
        contact.value.pincode = ''
    }

}
</script>

<template>
    <Navigation class="dark" />

    <div class="container mx-auto py-5 max-w-screen-md">
        <h1 class="font-nathejk text-4xl text-slate-700 pb-5">
            Hjælpertilmelding
        </h1>
            <div class="flex flex-col gap-2 mx-auto" style="min-height: 16rem; max-width: 20rem">
                <div class="mb-4">
                    <IconField>
                        <InputIcon><i class="pi pi-user" /></InputIcon>
                        <InputText v-model="person.name" type="text" placeholder="Navn" required />
                    </IconField>
                </div>
                <div class="mb-4">
                    <IconField>
                        <InputIcon><i class="pi pi-envelope" /></InputIcon>
                        <InputText v-model="person.email" type="email" placeholder="E-mail" required />
                    </IconField>
                </div>
                <div class="mb-4">
                    <IconField>
                        <InputIcon><i class="pi pi-mobile" /></InputIcon>
                        <InputText v-model="person.phone" type="text" placeholder="Telefonnummer" required />
                    </IconField>
                </div>
                <div class="flex flex-col">
                    <FloatLabel class="mt-7">
                        <Dropdown v-model="person.corps" inputId="member-korps" :options="config.corps" optionValue="slug" optionLabel="label"  class="w-full filled md:w-14rem" />
                        <label for="member-korps">Spejderkorps</label>
                    </FloatLabel>
                </div>
                <div class="mb-4">
                    <IconField>
                        <InputIcon><i class="pi pi-hashtag" /></InputIcon>
                        <InputText v-model="person.medlemNr" type="text" placeholder="Medlemsnummer" />
                    </IconField>
                </div>
                <div class="flex flex-col">
                    <FloatLabel class="mt-7">
                        <Dropdown v-model="person.department" inputId="member-area" :options="config.areas" optionValue="slug" optionLabel="label"  class="w-full filled md:w-14rem" />
                        <label for="member-area">Område på Nathejk</label>
                    </FloatLabel>
                </div>
                <div class="flex flex-col">
                    <FloatLabel class="mt-7">
                        <Dropdown v-model="person.diet" inputId="member-diet" :options="config.diets" optionValue="slug" optionLabel="label"  class="w-full filled md:w-14rem" />
                        <label for="member-diet">Spisevaner</label>
                    </FloatLabel>
                </div>
            </div>
            <div class="flex pt-4 justify-end">
                <Button label="Videre" icon="pi pi-arrow-right" iconPos="right" @click="save" :disabled="!complete"/>
            </div>

    </div>
</template>

<style>
</style>
