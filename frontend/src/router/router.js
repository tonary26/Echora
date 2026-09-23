import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/store/auth'

const router = createRouter({
    history: createWebHistory(),
    
    routes: [
        {
            path: "/",
            component: () => import("@/views/Daashboard.vue"),
            name: "dashboard",
            meta: { requireAuth: true }
        },
        {
            path: "/auth/register",
            component: () => import("@/views/auth/Register.vue"),
            name: "auth.register",
            meta: { requireAuth: false }
        },
        {
            path: "/auth/login",
            component: () => import("@/views/auth/Login.vue"),
            name: "auth.login",
            meta: { requireAuth: false }
        }
    ]
})

router.beforeEach((to) => {
  const authStore = useAuthStore()

  if (to.meta.requireAuth && !authStore.isAuthenticated) {
    return "/auth/login"
  } 
  if (!to.meta.requireAuth && authStore.isAuthenticated) {
    return "/"
  } 
})

export default router
