<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import Navigation from '@/components/Navigation.vue'

const router = useRouter()
const config = ref({ timeCountdown: "2024" });
const phone = ref(null);
const person = ref({phone:'', pincode:''});

async function loadConfig() {
    try {
        const response = await fetch("/api/home");
        if (!response.ok) {
            throw new Error("HTTP status " + response.status);
        }
        const data = await response.json();
        return data.config
    } catch (error) {
        console.log("Failed loading config", error);
    }
    return {}
}


onMounted(async () => {
    config.value = await loadConfig()
})


function link(url) {
    window.location.href=url
}
const signup = async () => {
    const headers = {
        "Content-Type": "application/json",
    }
    try {
        const body = JSON.stringify({
            phonePending: person.value.phone,
        })
        const response = await fetch("/api/start", { method: 'POST', body: body, headers: headers });
        if (!response.ok) {
            throw new Error("HTTP status " + response.status);
        }
        //const data = await response.json();
        //contact.value = data.team
        //router.replace({ name: 'indskrivning', params: { id: data.teamId } })
        //router.replace({ path: '/indskrivning/'+ data.team.teamId  })

        showpin.value = true
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
    try {
        const body = JSON.stringify({
            phonePending: person.value.phone,
            pincode: person.value.pincode,
        })
        const response = await fetch("/api/verify", { method: 'POST', body: body, headers: headers });
        if (!response.ok) {
            throw new Error("HTTP status " + response.status);
        }
        const data = await response.json();
        console.log("response data", data)
        router.replace({ path: '/indskrivning/'+ data.userId  })
        //router.replace({ path: data.team.teamPage })
    } catch (error) {
        console.log("team signup failed", error);
        person.value.pincode = ''
    }

}
</script>

<template>
  <header>
    <!-- Navbar -->
    <nav></nav>
    <!-- Navbar -->

    <!-- Add only the code below -->

    <!-- Background image -->
    <div class="h-screen bg-cover bg-no-repeat" style="background-image: url('/bg.jpg');">

        <div class=" h-screen overflow-hidden flex items-center justify-center">
  <div class="bg-white lg:w-6/12 md:7/12 w-8/12 shadow-3xl rounded-xl">
    <div class="bg-gray-800 shadow shadow-gray-200 absolute left-1/2 transform -translate-x-1/2 -translate-y-1/2 rounded-full p-4 md:p-8">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="#fff">
          <path d="M20.822 18.096c-3.439-.794-6.64-1.49-5.09-4.418 4.72-8.912 1.251-13.678-3.732-13.678-5.082 0-8.464 4.949-3.732 13.678 1.597 2.945-1.725 3.641-5.09 4.418-3.073.71-3.188 2.236-3.178 4.904l.004 1h23.99l.004-.969c.012-2.688-.092-4.222-3.176-4.935z"/>
        </svg>
    </div>
    <form v-if="!showpin" class="p-12 md:p-24" @submit.prevent="signup">
      <div class="flex items-center text-lg mb-6 md:mb-8">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="24" class="absolute ml-3">
  <path d="M10.5 18.75a.75.75 0 0 0 0 1.5h3a.75.75 0 0 0 0-1.5h-3Z" />
  <path fill-rule="evenodd" d="M8.625.75A3.375 3.375 0 0 0 5.25 4.125v15.75a3.375 3.375 0 0 0 3.375 3.375h6.75a3.375 3.375 0 0 0 3.375-3.375V4.125A3.375 3.375 0 0 0 15.375.75h-6.75ZM7.5 4.125C7.5 3.504 8.004 3 8.625 3H9.75v.375c0 .621.504 1.125 1.125 1.125h2.25c.621 0 1.125-.504 1.125-1.125V3h1.125c.621 0 1.125.504 1.125 1.125v15.75c0 .621-.504 1.125-1.125 1.125h-6.75A1.125 1.125 0 0 1 7.5 19.875V4.125Z" clip-rule="evenodd" />
</svg>

        <input type="text" id="phone" v-model="person.phone" class="bg-gray-200 rounded pl-12 py-2 md:py-4 focus:outline-none w-full" placeholder="Telefonnummer" />
      </div>
      <button class="bg-gradient-to-b from-gray-700 to-gray-900 font-medium p-2 md:p-4 text-white uppercase w-full rounded" :disabled="person.phone.length < 8">Send pinkode &raquo;</button>
    </form>
    <form v-else class="p-12 md:p-24" @submit.prevent="validatePincode">
      <div class="flex items-center text-lg mb-6 md:mb-8">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="24" class="absolute ml-3">
  <path d="M10.5 18.75a.75.75 0 0 0 0 1.5h3a.75.75 0 0 0 0-1.5h-3Z" />
  <path fill-rule="evenodd" d="M8.625.75A3.375 3.375 0 0 0 5.25 4.125v15.75a3.375 3.375 0 0 0 3.375 3.375h6.75a3.375 3.375 0 0 0 3.375-3.375V4.125A3.375 3.375 0 0 0 15.375.75h-6.75ZM7.5 4.125C7.5 3.504 8.004 3 8.625 3H9.75v.375c0 .621.504 1.125 1.125 1.125h2.25c.621 0 1.125-.504 1.125-1.125V3h1.125c.621 0 1.125.504 1.125 1.125v15.75c0 .621-.504 1.125-1.125 1.125h-6.75A1.125 1.125 0 0 1 7.5 19.875V4.125Z" clip-rule="evenodd" />
</svg>

        <input type="text" id="phone" v-model="person.phone" class="bg-gray-200 rounded pl-12 py-2 md:py-4 focus:outline-none w-full" :disabled="true" placeholder="Telefonnummer" />
      </div>
      <div v-if="showpin" class="flex items-center text-lg mb-6 md:mb-8">
        <svg class="absolute ml-3" viewBox="0 0 24 24" width="24">
          <path d="m18.75 9h-.75v-3c0-3.309-2.691-6-6-6s-6 2.691-6 6v3h-.75c-1.24 0-2.25 1.009-2.25 2.25v10.5c0 1.241 1.01 2.25 2.25 2.25h13.5c1.24 0 2.25-1.009 2.25-2.25v-10.5c0-1.241-1.01-2.25-2.25-2.25zm-10.75-3c0-2.206 1.794-4 4-4s4 1.794 4 4v3h-8zm5 10.722v2.278c0 .552-.447 1-1 1s-1-.448-1-1v-2.278c-.595-.347-1-.985-1-1.722 0-1.103.897-2 2-2s2 .897 2 2c0 .737-.405 1.375-1 1.722z"/>
        </svg>
        <input type="password" id="password" v-model="person.pincode" class="bg-gray-200 rounded pl-12 py-2 md:py-4 focus:outline-none w-full" placeholder="Pinkode fra SMS" />
      </div>
      <button class="bg-gradient-to-b from-gray-700 to-gray-900 font-medium p-2 md:p-4 text-white uppercase w-full rounded">Videre &raquo;</button>
    </form>
  </div>
 </div>
    </div>
    <!-- Background image -->

    <!-- Add only the code above -->
  </header>
  <main class="">

  </main>
  <footer v-if="false" class="bg-slate-900 text-slate-100 font-nathejk text-3xl text-center uppercase">
      <div class="py-52">Vi ses i mørket...</div>
  </footer>

</template>

<style>
.transition {
    transition: all 1s;
}

.transition.fading {
    opacity: 0;
}
</style>
