import { defineStore } from 'pinia'
import api from "@/api.js"

const readStoredUser = () => {
    try {
        return JSON.parse(localStorage.getItem("user"))
    } catch {
        localStorage.removeItem("user")
        return null
    }
}

export const useAuthStore = defineStore('auth', {
    state: () => ({
        user: readStoredUser(),
        token: localStorage.getItem("token")
    }),

    getters: {
        isAuthenticated: state => !!state.token,
    },

    actions: {
        async register(formData) {
            const { data } = await api.post("/auth/register", formData)
            localStorage.setItem("user", JSON.stringify(data.user))

            this.user = data.user
            return data
        },

        async login(formData) {
            const { data } = await api.post("/auth/login", formData)
            localStorage.setItem("token", data.token)
            localStorage.setItem("user", JSON.stringify(data.user))

            this.token = data.token
            this.user = data.user
            return data
        },

        async fetchCurrentUser() {
            const { data } = await api.get("/auth/me")
            localStorage.setItem("user", JSON.stringify(data.user))
            this.user = data.user
            return data.user
        },

        logout() {
            localStorage.removeItem("token")
            localStorage.removeItem("user")
            this.user = null
            this.token = null
        }
    },
})
