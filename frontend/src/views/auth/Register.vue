<script setup>
import { useAuthStore } from '@/store/auth'
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'

const formData = reactive({
    Name: "",
    Email: "",
    Avatar_url: "",
    Password: "",
})
const store = useAuthStore()
const router = useRouter()
const errorMessage = ref('')
const isSubmitting = ref(false)

const register = async function () {
    errorMessage.value = ''
    isSubmitting.value = true

    try {
      await store.register(formData)
      await store.login({ Email: formData.Email, Password: formData.Password })
      await router.push({ name: 'dashboard' })
    } catch (error) {
      errorMessage.value = error.response?.data?.error || 'Не удалось создать аккаунт. Попробуйте ещё раз.'
    } finally {
      isSubmitting.value = false
    }
}

</script>

<template>
  <div class="registration-page">
    <div class="sheet">

      <div class="intro">
        <div class="intro-copy">
          <h1>Начните вести свой архив</h1>
          <p>Один профиль хранит черновики, заметки и всё, что вы соберёте по пути.</p>
        </div>
        <div class="intro-foot">
          Уже есть профиль? <router-link :to="{ name: 'auth.login' }">Войти</router-link>
        </div>
        <div class="bars" aria-hidden="true">
          <i></i><i></i><i></i><i></i><i></i><i></i><i></i>
        </div>
      </div>

      <div class="form-side">
        <div class="form-head">
          <h2>Создание профиля</h2>
        </div>

        <form @submit.prevent="register">
          <div class="avatar-row">
            <div class="avatar-drop" v-if="formData.Avatar_url">
              <img :src="formData.Avatar_url" alt="Превью аватара" class="avatar-preview">
            </div>
            <div class="avatar-drop" v-else>
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                <circle cx="12" cy="8.5" r="3.2" stroke="currentColor" stroke-width="1.4"/>
                <path d="M5.5 19c1.2-3.2 3.9-4.8 6.5-4.8s5.3 1.6 6.5 4.8" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
              </svg>
            </div>
            <div class="avatar-text">
              <strong>Фото профиля</strong>
              <input
                v-model="formData.Avatar_url"
                type="url"
                placeholder="Ссылка на изображение"
                class="avatar-url-input"
              >
            </div>
          </div>

          <div class="field">
            <label for="name">Имя</label>
            <input v-model="formData.Name" type="text" id="name" name="name" placeholder="Как к вам обращаться" autocomplete="name">
          </div>

          <div class="field">
            <label for="email">Почта</label>
            <input v-model="formData.Email" type="email" id="email" name="email" placeholder="you@example.com" autocomplete="email">
          </div>

          <div class="field">
            <label for="password">Пароль</label>
            <input v-model="formData.Password" type="password" id="password" name="password" placeholder="Минимум 8 символов" autocomplete="new-password">
          </div>

          <p v-if="errorMessage" class="form-error" role="alert">{{ errorMessage }}</p>
          <button type="submit" class="submit" :disabled="isSubmitting">
            <span v-if="isSubmitting" class="button-spinner" aria-hidden="true"></span>
            {{ isSubmitting ? 'Создаём…' : 'Создать аккаунт' }}
          </button>
        </form>
      </div>

    </div>
  </div>
</template>

<style scoped>
.registration-page{
  --bg:#0E0E0E;
  --panel:#161616;
  --field:#212121;
  --ink:#FFFFFF;
  --ink-soft:#9C9C9C;
  --yellow:#FFCC00;
  --line:#2A2A2A;
  --radius-card:20px;
  --radius-field:12px;

  background:var(--bg);
  color:var(--ink);
  font-family:"Manrope", system-ui, sans-serif;
  min-height:100vh;
  width:100%;
  display:flex;
  align-items:center;
  justify-content:center;
  padding:clamp(20px, 4vw, 48px);
  box-sizing:border-box;
}

.registration-page[data-theme="light"]{
  --bg:#F3F3F1;
  --panel:#FFFFFF;
  --field:#F0F0EE;
  --ink:#111111;
  --ink-soft:#6B6B6B;
  --yellow:#FFCC00;
  --line:#E4E3DE;
}

.registration-page *{ box-sizing:border-box; }

.sheet{
  width:100%;
  max-width:940px;
  display:grid;
  grid-template-columns:0.85fr 1.15fr;
  background:var(--panel);
  border:1px solid var(--line);
  border-radius:var(--radius-card);
  overflow:hidden;
  min-height:600px;
  box-shadow:0 24px 70px rgba(0, 0, 0, 0.28);
}

.intro{
  background: #000000;
  color:#FFFFFF;
  padding:clamp(36px, 5vw, 52px) clamp(30px, 4vw, 44px);
  display:flex;
  flex-direction:column;
  justify-content:space-between;
  position:relative;
  overflow:hidden;
}

.intro-mark{
  display:flex;
  align-items:center;
  gap:8px;
  font-size:16px;
  font-weight:800;
  letter-spacing:0.01em;
}

.intro-mark .dot{
  width:9px; height:9px; border-radius:50%;
  background:var(--yellow);
  display:inline-block;
}

.intro-copy{ position:relative; z-index:1; }
.intro-copy h1{
  font-family:"Manrope", sans-serif;
  font-weight:800;
  font-size:clamp(30px, 4vw, 40px);
  line-height:1.1;
  letter-spacing:-0.01em;
  margin:0 0 16px;
  max-width:11ch;
}

.intro-copy p{
  font-size:15px;
  line-height:1.6;
  color:#B9B9B9;
  max-width:30ch;
  margin:0;
}

.intro-foot{
  font-size:13px;
  color:#8C8C8C;
  line-height:1.6;
  position:relative;
  z-index:1;
}

.intro-foot a{ color:var(--yellow); text-decoration:none; font-weight:600; }
.intro-foot a:hover{ text-decoration:underline; }

.bars{
  position:absolute;
  right:-30px;
  bottom:-20px;
  display:flex;
  align-items:flex-end;
  gap:6px;
  height:140px;
  opacity:0.9;
}

.bars i{
  display:block;
  width:9px;
  border-radius:5px;
  background:var(--yellow);
  opacity:0.16;
}

.bars i:nth-child(1){height:40%;}
.bars i:nth-child(2){height:70%;}
.bars i:nth-child(3){height:50%;}
.bars i:nth-child(4){height:95%;}
.bars i:nth-child(5){height:60%;}
.bars i:nth-child(6){height:80%;}
.bars i:nth-child(7){height:35%;}

.form-side{
  padding:clamp(36px, 5vw, 52px) clamp(30px, 4vw, 48px);
  display:flex;
  flex-direction:column;
  justify-content:center;
}

.form-head{ margin-bottom:26px; }
.form-head h2{
  font-weight:800;
  font-size:24px;
  letter-spacing:-0.01em;
  margin:0 0 6px;
}

.form-head p{
  margin:0;
  font-size:14px;
  color:var(--ink-soft);
}

form{ display:flex; flex-direction:column; gap:18px; }

.avatar-row{
  display:flex;
  align-items:center;
  gap:16px;
  margin-bottom:2px;
}

.avatar-drop{
  width:72px;
  height:72px;
  border-radius:50%;
  background:var(--field);
  border:1.5px dashed var(--line);
  display:flex;
  align-items:center;
  justify-content:center;
  flex-shrink:0;
  cursor:pointer;
  color:var(--ink-soft);
  transition:border-color .15s ease, color .15s ease;
  position:relative;
}

.avatar-preview{
  width:100%;
  height:100%;
  border-radius:50%;
  object-fit:cover;
}

.avatar-url-input{
  margin-top:4px;
  font-family:inherit;
  font-size:13px;
  padding:6px 10px;
  border:1.5px solid var(--line);
  border-radius:8px;
  background:var(--field);
  color:var(--ink);
  outline:none;
  width:100%;
  max-width:220px;
}

.avatar-url-input:focus{
  border-color:var(--yellow);
  box-shadow:0 0 0 3px rgba(255, 204, 0, 0.22);
}

.avatar-text strong{
  display:block;
  font-size:14px;
  font-weight:600;
  margin-bottom:2px;
}

.avatar-text span{
  font-size:12.5px;
  color:var(--ink-soft);
}

.field{ display:flex; flex-direction:column; gap:7px; }
.field label{
  font-size:13px;
  font-weight:600;
  color:var(--ink);
}

.field input[type="text"],
.field input[type="email"],
.field input[type="password"]{
  font-family:inherit;
  font-size:15px;
  padding:12px 14px;
  border:1.5px solid transparent;
  border-radius:var(--radius-field);
  background:var(--field);
  color:var(--ink);
  outline:none;
  transition:border-color .15s ease, box-shadow .15s ease;
  width:100%;
  min-height:48px;
}

.field input::placeholder{ color:var(--ink-soft); opacity:0.8; }
.field input:hover{ border-color:var(--line); }
.field input:focus{
  border-color:var(--yellow);
  box-shadow:0 0 0 3px rgba(255, 204, 0, 0.22);
}

.hint{ font-size:12px; color:var(--ink-soft); }

.submit{
  margin-top:6px;
  padding:14px 18px;
  background:var(--yellow);
  color:#111111;
  border:none;
  border-radius:999px;
  font-family:inherit;
  font-size:15px;
  font-weight:700;
  cursor:pointer;
  min-height:48px;
  transition:background .15s ease, transform .15s ease;
}

.submit:hover{ background:#FFD633; transform:translateY(-1px); }
.submit:active{ transform:translateY(0) scale(0.99); }
.submit:disabled{ cursor:wait; opacity:.72; transform:none; }
.form-error{
  margin:0;
  padding:11px 13px;
  border:1px solid rgba(255, 110, 94, .35);
  border-radius:10px;
  background:rgba(255, 110, 94, .08);
  color:#FFB7AF;
  font-size:13px;
  line-height:1.45;
}
.button-spinner{
  display:inline-block;
  width:14px;
  height:14px;
  margin-right:8px;
  border:2px solid rgba(17,17,17,.28);
  border-top-color:#111;
  border-radius:50%;
  vertical-align:-2px;
  animation:button-spin .7s linear infinite;
}
@keyframes button-spin{ to{ transform:rotate(360deg); } }
.submit:focus-visible,
.intro-foot a:focus-visible,
.divider-note a:focus-visible,
.avatar-drop:focus-within{
  outline:2px solid var(--yellow);
  outline-offset:2px;
}

.divider-note{
  text-align:center;
  font-size:12.5px;
  color:var(--ink-soft);
  margin-top:2px;
}

.divider-note a{ color:var(--yellow); text-decoration:none; font-weight:600; }
.divider-note a:hover{ text-decoration:underline; }

@media (max-width:720px){
  .registration-page{ padding:16px; align-items:flex-start; }
  .sheet{ grid-template-columns:1fr; min-height:auto; border-radius:16px; }
  .intro{ min-height:280px; padding:30px 24px; gap:48px; }
  .intro-copy h1{ max-width:none; }
  .form-side{ padding:32px 24px; }
}

@media (prefers-reduced-motion: reduce){
  .registration-page *{ transition:none !important; }
}
</style>
