<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { useTrackStore } from '@/store/tracks'

const router = useRouter()
const authStore = useAuthStore()
const trackStore = useTrackStore()
const { tracks, loading, uploading, uploadProgress, error } = storeToRefs(trackStore)

const audio = ref(null)
const fileInput = ref(null)
const coverInput = ref(null)
const uploadDialog = ref(null)
const currentTrack = ref(null)
const isPlaying = ref(false)
const hasStartedPlayback = ref(false)
const isBuffering = ref(false)
const isDragging = ref(false)
const uploadOpen = ref(false)
const uploadNotice = ref('')
const selectedAudio = ref(null)
const selectedCover = ref(null)
const coverPreview = ref('')
const currentTime = ref(0)
const duration = ref(0)
const volume = ref(0.82)
const avatarLoadFailed = ref(false)

const currentIndex = computed(() => tracks.value.findIndex(track => track.id === currentTrack.value?.id))
const avatarUrl = computed(() => authStore.user?.avatar_url?.trim() || '')
const progress = computed(() => duration.value ? (currentTime.value / duration.value) * 100 : 0)

const trackTitle = track => track.title.replace(/\.[^/.]+$/, '')
const trackExtension = track => track.title.split('.').pop()?.toUpperCase() || 'AUDIO'
const formatDate = value => new Intl.DateTimeFormat('ru', { day: 'numeric', month: 'short' }).format(new Date(value))

const formatTime = value => {
  if (!Number.isFinite(value)) return '0:00'
  const minutes = Math.floor(value / 60)
  return `${minutes}:${Math.floor(value % 60).toString().padStart(2, '0')}`
}

const handleUnauthorized = async requestError => {
  if (requestError?.response?.status !== 401) return false
  authStore.logout()
  await router.replace({ name: 'auth.login' })
  return true
}

const loadTracks = async () => {
  try {
    await trackStore.fetchTracks()
  } catch (requestError) {
    await handleUnauthorized(requestError)
  }
}

const resetUpload = () => {
	selectedAudio.value = null
	selectedCover.value = null
	if (coverPreview.value) URL.revokeObjectURL(coverPreview.value)
	coverPreview.value = ''
	if (fileInput.value) fileInput.value.value = ''
	if (coverInput.value) coverInput.value.value = ''
}

const openUpload = async () => {
	uploadOpen.value = true
	uploadNotice.value = ''
	trackStore.clearError()
	await nextTick()
	uploadDialog.value?.focus()
}

const closeUpload = () => {
	if (uploading.value) return
	uploadOpen.value = false
	uploadNotice.value = ''
	resetUpload()
}

const openFilePicker = () => {
  if (!uploading.value) fileInput.value?.click()
}

const selectAudio = file => {
	if (!file || uploading.value) return

  const audioExtension = /\.(mp3|wav|ogg|m4a|aac|flac|opus)$/i.test(file.name)
	if (!file.type.startsWith('audio/') && !audioExtension) {
    uploadNotice.value = 'Выберите аудиофайл: MP3, WAV, OGG, M4A, AAC, FLAC или OPUS.'
    return
  }
	if (file.size > 100 * 1024 * 1024) {
    uploadNotice.value = 'Файл больше 100 МБ. Выберите версию меньшего размера.'
    return
	}
	selectedAudio.value = file
	uploadNotice.value = ''
}

const selectCover = file => {
	if (!file || uploading.value) return
	if (!file.type.startsWith('image/')) {
		uploadNotice.value = 'Выберите изображение для обложки.'
		return
	}
	if (file.size > 10 * 1024 * 1024) {
		uploadNotice.value = 'Обложка больше 10 МБ. Выберите файл меньшего размера.'
		return
	}
	if (coverPreview.value) URL.revokeObjectURL(coverPreview.value)
	selectedCover.value = file
	coverPreview.value = URL.createObjectURL(file)
	uploadNotice.value = ''
}

const uploadFile = async () => {
	if (!selectedAudio.value || uploading.value) {
		uploadNotice.value = 'Сначала выберите аудиофайл.'
		return
	}

	uploadNotice.value = ''
	try {
		await trackStore.uploadTrack(selectedAudio.value, selectedCover.value)
		closeUpload()
	} catch (requestError) {
		await handleUnauthorized(requestError)
	}
}

const onFileChange = event => selectAudio(event.target.files?.[0])
const onCoverChange = event => selectCover(event.target.files?.[0])
const onDrop = event => {
	isDragging.value = false
	selectAudio(event.dataTransfer.files?.[0])
}

const playTrack = async track => {
  if (currentTrack.value?.id === track.id && audio.value?.src) {
    if (isPlaying.value) audio.value.pause()
    else await audio.value.play()
    return
  }

  currentTrack.value = track
  isBuffering.value = true
  trackStore.clearError()

  try {
    const url = await trackStore.getStreamUrl(track.id)
    currentTime.value = 0
    duration.value = 0
    await nextTick()
    audio.value.src = url
    audio.value.volume = volume.value
    audio.value.load()
    await audio.value.play()
  } catch (requestError) {
    await handleUnauthorized(requestError)
  } finally {
    isBuffering.value = false
  }
}

const togglePlayback = async () => {
  if (!currentTrack.value) {
    if (tracks.value[0]) await playTrack(tracks.value[0])
    return
  }
  if (isPlaying.value) audio.value.pause()
  else await audio.value.play()
}

const playAdjacent = direction => {
  if (!tracks.value.length) return
  const base = currentIndex.value < 0 ? 0 : currentIndex.value
  const nextIndex = (base + direction + tracks.value.length) % tracks.value.length
  playTrack(tracks.value[nextIndex])
}

const seek = event => {
  if (!audio.value || !duration.value) return
  audio.value.currentTime = (Number(event.target.value) / 100) * duration.value
}

const setVolume = event => {
  volume.value = Number(event.target.value)
  if (audio.value) audio.value.volume = volume.value
}

const onKeydown = event => {
	if (event.code === 'Escape' && uploadOpen.value) {
		closeUpload()
		return
	}
	if (event.code === 'Tab' && uploadOpen.value) {
		const focusable = [...uploadDialog.value.querySelectorAll('button:not(:disabled), [tabindex="0"]')]
		if (!focusable.length) return
		const first = focusable[0]
		const last = focusable[focusable.length - 1]
		if (event.shiftKey && document.activeElement === first) {
			event.preventDefault()
			last.focus()
		} else if (!event.shiftKey && document.activeElement === last) {
			event.preventDefault()
			first.focus()
		}
		return
	}
	const tag = document.activeElement?.tagName
  if (event.code !== 'Space' || ['INPUT', 'BUTTON', 'TEXTAREA'].includes(tag)) return
  event.preventDefault()
  togglePlayback()
}

const logout = async () => {
  audio.value?.pause()
  authStore.logout()
  await router.push({ name: 'auth.login' })
}

onMounted(async () => {
  window.addEventListener('keydown', onKeydown)

  try {
    await Promise.all([authStore.fetchCurrentUser(), trackStore.fetchTracks()])
  } catch (requestError) {
    await handleUnauthorized(requestError)
  }
})

watch(uploadOpen, isOpen => {
	document.body.style.overflow = isOpen ? 'hidden' : ''
})

watch(avatarUrl, () => {
  avatarLoadFailed.value = false
})

onBeforeUnmount(() => {
	window.removeEventListener('keydown', onKeydown)
	document.body.style.overflow = ''
	if (coverPreview.value) URL.revokeObjectURL(coverPreview.value)
})
</script>

<template>
  <div class="dashboard-shell" :class="{ 'has-player': hasStartedPlayback }">
    <aside class="sidebar">
      <nav aria-label="Основная навигация">
        <button class="nav-item active" type="button">
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 10.5 12 4l8 6.5V20H4v-9.5Z"/><path d="M9 20v-6h6v6"/></svg>
          Библиотека
        </button>
      </nav>
      <div class="sidebar-foot">
        <div class="profile-mark">
          <img v-if="avatarUrl && !avatarLoadFailed" :src="avatarUrl" :alt="`Аватар пользователя ${authStore.user?.name || ''}`" @error="avatarLoadFailed = true">
          <span v-else aria-hidden="true">{{ authStore.user?.name?.charAt(0)?.toUpperCase() || 'E' }}</span>
        </div>
        <div class="profile-copy">
          <strong>{{ authStore.user?.name || 'Мой профиль' }}</strong>
          <span>{{ authStore.user?.email || 'Личная коллекция' }}</span>
        </div>
        <button class="icon-button" type="button" aria-label="Выйти" title="Выйти" @click="logout">
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M10 5H5v14h5M14 8l4 4-4 4m4-4H9"/></svg>
        </button>
      </div>
    </aside>

    <main class="content">
      <header class="page-header">
        <div>
          <h1>Ваша музыка.<br><em>Без лишнего шума.</em></h1>
          <p>Личная аудиотека: загружайте записи и продолжайте слушать с любого трека.</p>
        </div>
        <button class="upload-button" type="button" aria-haspopup="dialog" @click="openUpload">
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 17V5m0 0L7.5 9.5M12 5l4.5 4.5"/><path d="M5 15v4h14v-4"/></svg>
          Добавить трек
        </button>
      </header>

      <section class="library" aria-labelledby="library-title">
        <div class="section-heading">
          <div><h2 id="library-title">Библиотека</h2><p v-if="loading">Синхронизируем коллекцию…</p></div>
          <button class="refresh-button" type="button" :disabled="loading" @click="loadTracks">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M19 8a7 7 0 1 0 .2 7.5M19 4v4h-4"/></svg> Обновить
          </button>
        </div>

        <div v-if="error" class="error-banner" role="alert"><span>{{ error }}</span><button type="button" @click="loadTracks">Повторить</button></div>

        <div v-if="loading" class="skeleton-list" aria-label="Загрузка треков">
          <div v-for="item in 4" :key="item" class="skeleton-row"><i></i><span></span><small></small></div>
        </div>

        <div v-else-if="!tracks.length" class="empty-state">
          <div class="empty-wave" aria-hidden="true"><i></i><i></i><i></i><i></i><i></i></div>
          <h3>Здесь пока тихо</h3>
          <p>Добавьте первую запись. Она появится здесь и сразу будет готова к прослушиванию.</p>
          <button type="button" @click="openUpload">Загрузить первый трек</button>
        </div>

        <TransitionGroup v-else name="track-list" tag="ol" class="track-list">
          <li v-for="(track, index) in tracks" :key="track.id" class="track-row"
            :class="{ current: currentTrack?.id === track.id }"
            :style="{ '--delay': `${Math.min(index, 7) * 35}ms` }">
            <button class="track-main" type="button" @click="playTrack(track)">
              <span class="track-art" :class="{ 'vinyl-fallback': !track.cover_url }" aria-hidden="true">
                <img v-if="track.cover_url" :src="track.cover_url" alt="">
                <span class="art-ring"></span>
                <span v-if="currentTrack?.id === track.id && isPlaying" class="equalizer"><i></i><i></i><i></i></span>
                <svg v-else viewBox="0 0 24 24"><path d="m9 7 9 5-9 5V7Z"/></svg>
              </span>
              <span class="track-copy"><strong>{{ trackTitle(track) }}</strong><span>{{ trackExtension(track) }} · {{ formatDate(track.created_at) }}</span></span>
            </button>
            <button class="row-play" type="button" :aria-label="`Воспроизвести ${trackTitle(track)}`" @click="playTrack(track)">
              <span v-if="isBuffering && currentTrack?.id === track.id" class="mini-spinner"></span>
              <svg v-else viewBox="0 0 24 24" aria-hidden="true">
                <path v-if="currentTrack?.id === track.id && isPlaying" d="M8 6h3v12H8zM14 6h3v12h-3z"/>
                <path v-else d="m9 7 9 5-9 5V7Z"/>
              </svg>
            </button>
          </li>
        </TransitionGroup>
      </section>
    </main>

    <audio ref="audio" preload="metadata" hidden @play="isPlaying = true; hasStartedPlayback = true" @pause="isPlaying = false"
      @waiting="isBuffering = true" @canplay="isBuffering = false"
      @timeupdate="currentTime = audio?.currentTime || 0" @durationchange="duration = audio?.duration || 0" @ended="playAdjacent(1)"></audio>

    <footer v-if="hasStartedPlayback" class="player" :class="{ active: currentTrack, playing: isPlaying }">
      <div class="now-playing">
        <div class="player-art" :class="{ 'vinyl-fallback': !currentTrack?.cover_url }" aria-hidden="true">
          <img v-if="currentTrack?.cover_url" :src="currentTrack.cover_url" alt="">
          <span></span>
        </div>
        <div class="player-copy"><strong>{{ currentTrack ? trackTitle(currentTrack) : 'Выберите трек' }}</strong><span>{{ currentTrack ? trackExtension(currentTrack) : 'Ваша библиотека готова' }}</span></div>
      </div>
      <div class="transport">
        <div class="controls">
          <button type="button" aria-label="Предыдущий трек" :disabled="!tracks.length" @click="playAdjacent(-1)"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 5v14M18 6l-9 6 9 6V6Z"/></svg></button>
          <button class="play-button" type="button" :aria-label="isPlaying ? 'Пауза' : 'Воспроизвести'" :disabled="!tracks.length || isBuffering" @click="togglePlayback">
            <span v-if="isBuffering" class="mini-spinner dark"></span>
            <svg v-else viewBox="0 0 24 24" aria-hidden="true"><path v-if="isPlaying" d="M8 6h3v12H8zM14 6h3v12h-3z"/><path v-else d="m9 7 9 5-9 5V7Z"/></svg>
          </button>
          <button type="button" aria-label="Следующий трек" :disabled="!tracks.length" @click="playAdjacent(1)"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M18 5v14M6 6l9 6-9 6V6Z"/></svg></button>
        </div>
        <div class="timeline"><span>{{ formatTime(currentTime) }}</span><input type="range" min="0" max="100" step="0.1" :value="progress" :style="{ '--progress': `${progress}%` }" aria-label="Позиция трека" @input="seek"><span>{{ formatTime(duration) }}</span></div>
      </div>
      <div class="volume"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 10v4h4l5 4V6l-5 4H5Zm12-1.5a5 5 0 0 1 0 7"/></svg><input type="range" min="0" max="1" step="0.01" :value="volume" :style="{ '--progress': `${volume * 100}%` }" aria-label="Громкость" @input="setVolume"></div>
    </footer>

    <Teleport to="body">
      <Transition name="modal-fade">
        <div v-if="uploadOpen" class="modal-backdrop" @mousedown.self="closeUpload">
          <section ref="uploadDialog" class="upload-modal" role="dialog" aria-modal="true" aria-labelledby="upload-heading" tabindex="-1">
            <header class="modal-header">
              <div>
                <h2 id="upload-heading">Добавить трек</h2>
                <p>Выберите аудиофайл и, если хотите, добавьте обложку.</p>
              </div>
              <button class="modal-close" type="button" aria-label="Закрыть окно" :disabled="uploading" @click="closeUpload">
                <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m6 6 12 12M18 6 6 18"/></svg>
              </button>
            </header>

            <form @submit.prevent="uploadFile">
              <div class="upload-fields">
                <div class="drop-zone" :class="{ dragging: isDragging, uploading, selected: selectedAudio }" role="button" tabindex="0"
                  @click="openFilePicker" @keydown.enter="openFilePicker" @keydown.space.prevent="openFilePicker"
                  @dragover.prevent="isDragging = true" @dragleave.prevent="isDragging = false" @drop.prevent="onDrop">
                  <input ref="fileInput" type="file" accept="audio/*,.flac,.opus" hidden @change="onFileChange">
                  <div class="upload-glyph" aria-hidden="true">
                    <svg v-if="selectedAudio" viewBox="0 0 24 24"><path d="m5 12 4 4L19 6"/></svg>
                    <svg v-else viewBox="0 0 24 24"><path d="M12 16V4m0 0L7.5 8.5M12 4l4.5 4.5"/><path d="M5 14v5h14v-5"/></svg>
                  </div>
                  <div class="file-copy">
                    <strong>{{ selectedAudio ? selectedAudio.name : 'Перетащите аудиофайл сюда' }}</strong>
                    <span>{{ selectedAudio ? 'Нажмите, чтобы выбрать другой файл' : 'или нажмите, чтобы выбрать, до 100 МБ' }}</span>
                  </div>
                  <div v-if="uploading" class="upload-progress" aria-hidden="true"><i :style="{ transform: `scaleX(${uploadProgress / 100})` }"></i></div>
                </div>

                <button class="cover-picker" type="button" :disabled="uploading" @click="coverInput?.click()">
                  <input ref="coverInput" type="file" accept="image/*" hidden @change="onCoverChange">
                  <span class="cover-preview" aria-hidden="true">
                    <img v-if="coverPreview" :src="coverPreview" alt="">
                    <svg v-else viewBox="0 0 24 24"><path d="M4 5h16v14H4z"/><path d="m4 16 4.5-4.5 3 3 2-2L20 18M15.5 9h.01"/></svg>
                  </span>
                  <span><strong>{{ selectedCover ? selectedCover.name : 'Добавить обложку' }}</strong><small>Изображение до 10 МБ, необязательно</small></span>
                </button>
              </div>

              <p v-if="uploadNotice || error" class="notice" role="alert">{{ uploadNotice || error }}</p>
              <div class="modal-actions">
                <button class="cancel-button" type="button" :disabled="uploading" @click="closeUpload">Отмена</button>
                <button class="submit-button" type="submit" :disabled="!selectedAudio || uploading">
                  {{ uploading ? `Загружаем ${uploadProgress}%` : 'Добавить в библиотеку' }}
                </button>
              </div>
            </form>
          </section>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.dashboard-shell{--bg:#0b0b0b;--panel:#121212;--field:#202020;--ink:#f8f7f2;--muted:#9c9a92;--yellow:#ffcc00;--line:#2b2b28;min-height:100vh;display:grid;grid-template-columns:240px minmax(0,1fr);background:var(--bg);color:var(--ink)}.dashboard-shell.has-player{padding-bottom:104px}
.dashboard-shell::before{content:"";position:fixed;inset:0;pointer-events:none;opacity:.035;background-image:url("data:image/svg+xml,%3Csvg viewBox='0 0 180 180' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.9' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='.65'/%3E%3C/svg%3E");z-index:20}
.dashboard-shell *{box-sizing:border-box}button,input{font:inherit}button{color:inherit}svg{width:22px;height:22px;fill:none;stroke:currentColor;stroke-width:1.7;stroke-linecap:round;stroke-linejoin:round}
.sidebar{position:fixed;inset:0 auto 0 0;width:240px;padding:30px 22px 22px;display:flex;flex-direction:column;background:#050505;border-right:1px solid var(--line);z-index:5}.dashboard-shell.has-player .sidebar{bottom:104px}
nav{display:grid;gap:7px}.nav-item{width:100%;display:flex;align-items:center;gap:12px;padding:12px;border:0;border-radius:9px;background:transparent;color:var(--muted);cursor:pointer;text-align:left;font-size:14px;font-weight:650;transition:background .18s ease,color .18s ease,transform .18s ease}.nav-item:hover{background:var(--field);color:var(--ink);transform:translateX(2px)}.nav-item.active{background:var(--yellow);color:#111}.nav-item svg{width:19px;height:19px}
.sidebar-foot{margin-top:auto;display:grid;grid-template-columns:38px minmax(0,1fr) 34px;gap:10px;align-items:center;padding-top:20px;border-top:1px solid var(--line)}.profile-mark{width:38px;height:38px;border-radius:50%;display:grid;place-items:center;overflow:hidden;background:var(--field);color:var(--yellow);font-weight:800}.profile-mark img{width:100%;height:100%;display:block;object-fit:cover}.profile-copy{min-width:0;display:grid;gap:2px}.profile-copy strong,.profile-copy span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.profile-copy strong{font-size:13px}.profile-copy span{color:var(--muted);font-size:11px}.icon-button,.refresh-button,.row-play,.controls button{border:0;background:transparent;cursor:pointer}.icon-button{width:34px;height:34px;display:grid;place-items:center;border-radius:50%;color:var(--muted)}.icon-button:hover{background:var(--field);color:var(--ink)}.icon-button svg{width:18px}
.content{grid-column:2;padding:clamp(32px,5vw,72px) clamp(26px,6vw,86px) 56px;max-width:1320px;width:100%;margin:0 auto}.page-header{display:flex;justify-content:space-between;align-items:flex-end;gap:32px;margin-bottom:52px}.page-header h1{margin:0;max-width:680px;font-size:clamp(42px,6vw,78px);line-height:.96;letter-spacing:-.04em;font-weight:820;text-wrap:balance}.page-header h1 em{color:var(--yellow);font-style:normal}.page-header p{max-width:570px;margin:22px 0 0;color:var(--muted);font-size:15px;line-height:1.65}
.upload-button,.empty-state button{display:flex;align-items:center;justify-content:center;gap:9px;flex:none;border:0;border-radius:999px;padding:13px 20px;background:var(--yellow);color:#111;font-size:14px;font-weight:750;cursor:pointer;box-shadow:0 8px 24px rgba(255,204,0,.13);transition:transform .2s cubic-bezier(.16,1,.3,1),box-shadow .2s ease,background .2s ease}.upload-button:hover,.empty-state button:hover{transform:translateY(-2px);background:#ffd632;box-shadow:0 12px 30px rgba(255,204,0,.2)}.upload-button svg{width:18px}
.modal-backdrop{--yellow:#ffcc00;position:fixed;z-index:40;inset:0;display:grid;place-items:center;padding:24px;background:rgba(0,0,0,.76)}.upload-modal{width:min(620px,100%);max-height:calc(100vh - 48px);overflow:auto;border:1px solid #34332f;border-radius:18px;background:#121212;color:#f8f7f2;box-shadow:0 28px 80px rgba(0,0,0,.55);outline:none}.modal-header{display:flex;align-items:flex-start;justify-content:space-between;gap:20px;padding:24px 24px 20px;border-bottom:1px solid #2b2b28}.modal-header h2{margin:0;font-size:25px;letter-spacing:-.025em}.modal-header p{margin:7px 0 0;color:#9c9a92;font-size:13px;line-height:1.55}.modal-close{width:38px;height:38px;display:grid;place-items:center;flex:none;border:0;border-radius:50%;background:#202020;color:#9c9a92;cursor:pointer}.modal-close:hover:not(:disabled){color:#f8f7f2;background:#292929}.modal-close:disabled{opacity:.45;cursor:not-allowed}.modal-close svg{width:19px}.upload-modal form{padding:24px}.upload-fields{display:grid;gap:12px}.drop-zone{position:relative;min-height:156px;padding:26px;display:flex;align-items:center;gap:18px;border:1px dashed #55534b;border-radius:14px;background:#181818;cursor:pointer;overflow:hidden;transition:border-color .2s ease,background-color .2s ease}.drop-zone:hover,.drop-zone.dragging{border-color:#ffcc00;background:#1c1b16}.drop-zone.selected{border-style:solid;border-color:#666040}.drop-zone.uploading{cursor:wait}.upload-glyph{width:58px;height:58px;display:grid;place-items:center;flex:none;border-radius:50%;background:#ffcc00;color:#111}.upload-glyph svg{width:25px;height:25px}.file-copy{min-width:0;display:grid;gap:6px}.file-copy strong{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:16px;letter-spacing:-.01em}.file-copy span{color:#9c9a92;font-size:12px}.upload-progress{position:absolute;inset:auto 0 0;height:4px;background:#26251f}.upload-progress i{display:block;width:100%;height:100%;transform-origin:left;background:#ffcc00;transition:transform .2s ease}.cover-picker{width:100%;display:flex;align-items:center;gap:14px;padding:10px;border:1px solid #34332f;border-radius:12px;background:#181818;color:#f8f7f2;text-align:left;cursor:pointer;transition:border-color .2s ease,background-color .2s ease}.cover-picker:hover:not(:disabled){border-color:#55534b;background:#1d1d1d}.cover-picker:disabled{opacity:.55;cursor:wait}.cover-preview{width:58px;height:58px;display:grid;place-items:center;flex:none;overflow:hidden;border-radius:8px;background:#242424;color:#9c9a92}.cover-preview img{width:100%;height:100%;object-fit:cover}.cover-preview svg{width:23px}.cover-picker>span:last-child{min-width:0;display:grid;gap:4px}.cover-picker strong{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:14px}.cover-picker small{color:#9c9a92;font-size:11px}.notice{margin:12px 2px 0;color:#ffb7af;font-size:13px}.modal-actions{display:flex;justify-content:flex-end;gap:10px;margin-top:22px}.cancel-button,.submit-button{min-height:43px;padding:0 17px;border-radius:999px;font-size:13px;font-weight:750;cursor:pointer}.cancel-button{border:1px solid #3b3b37;background:transparent;color:#d7d5ca}.cancel-button:hover:not(:disabled){border-color:#66645e;color:#fff}.submit-button{border:0;background:#ffcc00;color:#111}.submit-button:hover:not(:disabled){background:#ffd632}.cancel-button:disabled,.submit-button:disabled{opacity:.45;cursor:not-allowed}.modal-fade-enter-active{transition:opacity .22s ease}.modal-fade-leave-active{transition:opacity .16s ease}.modal-fade-enter-active .upload-modal{transition:opacity .22s ease,transform .38s cubic-bezier(.16,1,.3,1),filter .3s ease}.modal-fade-leave-active .upload-modal{transition:opacity .16s ease,transform .16s ease}.modal-fade-enter-from,.modal-fade-leave-to{opacity:0}.modal-fade-enter-from .upload-modal{opacity:0;transform:translateY(18px) scale(.98);filter:blur(4px)}.modal-fade-leave-to .upload-modal{opacity:0;transform:translateY(8px) scale(.99)}
.section-heading{display:flex;align-items:flex-end;justify-content:space-between;gap:20px;margin-bottom:18px}.section-heading h2{margin:0;font-size:25px;letter-spacing:-.025em}.section-heading p{margin:5px 0 0;color:var(--muted);font-size:13px}.refresh-button{display:flex;align-items:center;gap:8px;padding:8px 2px;color:var(--muted);font-size:13px}.refresh-button:hover{color:var(--ink)}.refresh-button:disabled{opacity:.45;cursor:wait}.refresh-button:disabled svg{animation:spin .8s linear infinite}.refresh-button svg{width:17px}@keyframes spin{to{transform:rotate(360deg)}}
.error-banner{display:flex;align-items:center;justify-content:space-between;gap:16px;margin:0 0 14px;padding:12px 14px;border:1px solid rgba(255,110,94,.35);border-radius:10px;background:rgba(255,110,94,.08);color:#ffb7af;font-size:13px}.error-banner button{border:0;background:transparent;color:#ffd2cd;text-decoration:underline;text-underline-offset:3px;cursor:pointer}
.track-list{list-style:none;margin:0;padding:0;border-top:1px solid var(--line)}.track-row{display:grid;grid-template-columns:minmax(0,1fr) 40px 46px;align-items:center;gap:14px;min-height:78px;border-bottom:1px solid var(--line);transition:background .18s ease,border-color .18s ease;animation:row-arrive .48s cubic-bezier(.16,1,.3,1) both;animation-delay:var(--delay)}@keyframes row-arrive{from{opacity:0;transform:translateY(12px);filter:blur(3px)}to{opacity:1;transform:none;filter:none}}.track-row:hover,.track-row.current{background:#141414}.track-row.current{border-color:#464127}.track-main{min-width:0;display:flex;align-items:center;gap:15px;padding:10px 12px 10px 8px;border:0;background:transparent;color:inherit;text-align:left;cursor:pointer}.track-art,.player-art{position:relative;overflow:hidden;background:hsl(var(--hue) 74% 48%);color:#111}.track-art{width:52px;height:52px;display:grid;place-items:center;flex:none;border-radius:7px;box-shadow:0 8px 24px hsl(var(--hue) 70% 35% / .12)}.track-art::before,.player-art::before{content:"";position:absolute;inset:-30%;background:repeating-linear-gradient(70deg,transparent 0 9px,rgba(255,255,255,.18) 10px 11px);transform:rotate(14deg)}.track-art svg{position:relative;width:20px;fill:currentColor;stroke:none;transition:transform .2s ease}.track-main:hover .track-art svg{transform:scale(1.13)}.art-ring{position:absolute;width:30px;height:30px;border:1px solid rgba(10,10,10,.25);border-radius:50%}.equalizer{position:relative;height:19px;display:flex;align-items:flex-end;gap:3px}.equalizer i{width:3px;border-radius:2px;background:#111;animation:eq .7s ease-in-out infinite alternate}.equalizer i:nth-child(1){height:45%}.equalizer i:nth-child(2){height:95%;animation-delay:-.3s}.equalizer i:nth-child(3){height:65%;animation-delay:-.5s}@keyframes eq{to{height:25%}}.track-copy{min-width:0;display:grid;gap:5px}.track-copy strong{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:15px;letter-spacing:-.01em}.track-copy span{color:var(--muted);font-size:11px;letter-spacing:.06em}.track-number{color:#66645e;font-size:12px;font-variant-numeric:tabular-nums}.row-play{width:36px;height:36px;display:grid;place-items:center;border-radius:50%;color:var(--muted);transition:background .18s ease,color .18s ease,transform .18s ease}.row-play:hover{background:var(--yellow);color:#111;transform:scale(1.05)}.row-play svg{width:17px;fill:currentColor;stroke:none}.track-list-enter-active,.track-list-leave-active{transition:opacity .25s ease,transform .35s cubic-bezier(.16,1,.3,1)}.track-list-enter-from,.track-list-leave-to{opacity:0;transform:translateX(18px)}
.track-art>img,.player-art>img{position:absolute;inset:0;width:100%;height:100%;object-fit:cover}.track-art:has(img)::before,.player-art:has(img)::before{display:none}.track-art:has(img) .art-ring{display:none}.track-art:has(img) .equalizer i{background:#fff;box-shadow:0 1px 5px rgba(0,0,0,.65)}
.skeleton-list{border-top:1px solid var(--line)}.skeleton-row{height:78px;display:grid;grid-template-columns:52px minmax(0,320px) 60px;align-items:center;gap:16px;border-bottom:1px solid var(--line)}.skeleton-row i,.skeleton-row span,.skeleton-row small{display:block;border-radius:6px;background:linear-gradient(90deg,#171717 20%,#242424 45%,#171717 70%);background-size:250% 100%;animation:shimmer 1.25s linear infinite}.skeleton-row i{width:52px;height:52px}.skeleton-row span{height:14px}.skeleton-row small{height:10px}@keyframes shimmer{to{background-position:-250% 0}}.empty-state{min-height:320px;display:grid;place-items:center;align-content:center;text-align:center;border:1px dashed var(--line);border-radius:16px;padding:40px}.empty-state h3{margin:18px 0 8px;font-size:24px}.empty-state p{max-width:450px;margin:0 0 22px;color:var(--muted);line-height:1.6;font-size:14px}.empty-wave{height:48px;display:flex;align-items:center;gap:5px}.empty-wave i{display:block;width:5px;border-radius:4px;background:var(--yellow);animation:quiet-wave 1.2s ease-in-out infinite alternate}.empty-wave i:nth-child(1),.empty-wave i:nth-child(5){height:18px}.empty-wave i:nth-child(2),.empty-wave i:nth-child(4){height:34px;animation-delay:-.4s}.empty-wave i:nth-child(3){height:48px;animation-delay:-.8s}@keyframes quiet-wave{to{transform:scaleY(.45);opacity:.45}}
.player{position:fixed;z-index:10;inset:auto 0 0;height:104px;display:grid;grid-template-columns:1fr minmax(320px,600px) 1fr;align-items:center;gap:28px;padding:14px 28px;background:rgba(8,8,8,.96);border-top:1px solid var(--line);box-shadow:0 -18px 48px rgba(0,0,0,.28);transition:border-color .3s ease,background .3s ease}.player.active{border-color:#464127;background:rgba(10,10,8,.98)}.now-playing{min-width:0;display:flex;align-items:center;gap:12px}.player-art{width:58px;height:58px;flex:none;border-radius:8px;transition:border-radius 1.15s cubic-bezier(.16,1,.3,1)}.player-art span{position:absolute;inset:19px;border:3px solid rgba(10,10,10,.62);border-radius:50%;background:rgba(255,255,255,.45)}.player-art:has(img) span{border-color:rgba(255,255,255,.74);background:rgba(10,10,10,.55)}.player.playing .player-art{animation:cover-turn 8s linear infinite;border-radius:50%}@keyframes cover-turn{to{transform:rotate(360deg)}}.player-copy{min-width:0;display:grid;gap:4px}.player-copy strong,.player-copy span{white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.player-copy strong{font-size:14px}.player-copy span{color:var(--muted);font-size:11px;letter-spacing:.05em}.transport{display:grid;gap:9px}.controls{display:flex;justify-content:center;align-items:center;gap:12px}.controls button{width:34px;height:34px;display:grid;place-items:center;border-radius:50%;color:var(--muted)}.controls button:hover:not(:disabled){color:var(--ink);background:var(--field)}.controls button:disabled{opacity:.3;cursor:not-allowed}.controls .play-button{width:42px;height:42px;background:var(--yellow);color:#111;box-shadow:0 7px 20px rgba(255,204,0,.16);transition:transform .15s ease,background .15s ease}.controls .play-button:hover:not(:disabled){background:#ffd632;color:#111;transform:scale(1.06)}.controls svg{width:19px;fill:currentColor;stroke:none}.timeline{display:grid;grid-template-columns:38px 1fr 38px;align-items:center;gap:10px;color:#77756e;font-size:10px;font-variant-numeric:tabular-nums}.timeline span:last-child{text-align:right}input[type="range"]{--progress:0%;width:100%;height:3px;margin:0;appearance:none;border-radius:4px;outline:none;background:linear-gradient(to right,var(--yellow) 0 var(--progress),#34332f var(--progress) 100%);cursor:pointer}input[type="range"]::-webkit-slider-thumb{width:11px;height:11px;appearance:none;border:0;border-radius:50%;background:var(--ink);box-shadow:0 2px 8px rgba(0,0,0,.4)}input[type="range"]::-moz-range-thumb{width:11px;height:11px;border:0;border-radius:50%;background:var(--ink)}.volume{justify-self:end;width:min(160px,100%);display:flex;align-items:center;gap:10px;color:var(--muted)}.volume svg{width:18px}.volume input{max-width:110px}.mini-spinner{width:16px;height:16px;border:2px solid rgba(255,255,255,.22);border-top-color:var(--yellow);border-radius:50%;animation:spin .7s linear infinite}.mini-spinner.dark{border-color:rgba(17,17,17,.2);border-top-color:#111}
.sidebar-foot{padding-top:0;border-top:0}.track-row{grid-template-columns:minmax(0,1fr) 46px}.track-art,.player-art{border-radius:12px;background:#0a0a0a}.track-art{box-shadow:0 8px 24px rgba(0,0,0,.28)}.track-art.vinyl-fallback::before,.player-art.vinyl-fallback::before{inset:3px;border-radius:50%;transform:none;background:radial-gradient(circle,#171717 0 4%,var(--yellow) 5% 20%,#0a0a0a 21% 28%,transparent 29%),repeating-radial-gradient(circle,#151515 0 2px,#090909 3px 5px);box-shadow:inset 0 0 0 1px rgba(255,255,255,.06),0 5px 14px rgba(0,0,0,.45)}.track-art.vinyl-fallback .art-ring{display:none}.track-art.vinyl-fallback svg,.track-art.vinyl-fallback .equalizer{z-index:1}.player-art.vinyl-fallback span{inset:50% auto auto 50%;width:16px;height:16px;transform:translate(-50%,-50%);border:0;background:var(--yellow);box-shadow:inset 0 0 0 5px var(--yellow),inset 0 0 0 6px #171717;z-index:1}
button:focus-visible,input:focus-visible,.drop-zone:focus-visible{outline:2px solid var(--yellow);outline-offset:3px}::selection{background:var(--yellow);color:#111}
@media(max-width:860px){.dashboard-shell{display:block;padding:0 0 156px}.sidebar{position:sticky;inset:0 0 auto;width:100%;height:68px;padding:0 18px;display:flex;flex-direction:row;align-items:center;border-right:0;border-bottom:1px solid var(--line)}.brand{margin:0 auto 0 0}.brand-pulse{animation:none}.sidebar nav{display:flex}.nav-item{width:auto;padding:10px}.nav-item svg{display:none}.nav-item span{margin-left:6px}.sidebar-foot{margin:0 0 0 10px;padding:0 0 0 14px;border-top:0;border-left:1px solid var(--line);display:flex}.profile-copy,.profile-mark{display:none}.content{padding:38px 20px 36px}.page-header{align-items:flex-start;margin-bottom:38px}.page-header h1{font-size:clamp(38px,10vw,62px)}.player{height:140px;grid-template-columns:1fr auto;grid-template-rows:auto auto;padding:12px 18px;gap:10px 16px}.transport{grid-column:1/-1;grid-row:2}.volume{display:none}.now-playing{min-width:0}.player::after{content:"Пробел: пауза";justify-self:end;color:#5f5e58;font-size:10px}}
@media(max-width:860px){.dashboard-shell:not(.has-player){padding-bottom:0}}
@media(max-width:860px){.sidebar-foot{margin-left:auto;padding-left:0;border-left:0}}
@media(max-width:560px){.sidebar nav .nav-item:first-child{display:none}.page-header{display:grid}.page-header h1{font-size:42px}.page-header p{margin-top:16px}.upload-button{justify-self:start}.modal-backdrop{padding:10px}.upload-modal{max-height:calc(100vh - 20px);border-radius:14px}.modal-header,.upload-modal form{padding:18px}.drop-zone{min-height:176px;display:grid;text-align:center;justify-items:center;padding:22px}.file-copy{max-width:100%}.modal-actions{display:grid;grid-template-columns:1fr 1fr}.track-row{grid-template-columns:minmax(0,1fr) 42px}.track-number{display:none}.content{padding-inline:14px}.section-heading{padding-inline:4px}.refresh-button{font-size:0}.refresh-button svg{width:20px}.player{grid-template-columns:minmax(0,1fr) auto}.player::after{content:""}.player-art{width:46px;height:46px}.player-art span{inset:15px}.timeline{grid-template-columns:32px 1fr 32px}.empty-state{min-height:280px;padding:28px 20px}}
@media(prefers-reduced-motion:reduce){.brand-pulse,.equalizer i,.empty-wave i,.player.playing .player-art,.skeleton-row i,.skeleton-row span,.skeleton-row small,.refresh-button:disabled svg,.mini-spinner{animation-duration:1ms;animation-iteration-count:1}.track-row{animation:none}.modal-fade-enter-active,.modal-fade-leave-active,.modal-fade-enter-active .upload-modal,.modal-fade-leave-active .upload-modal,.track-list-enter-active,.track-list-leave-active{transition:opacity .15s ease}.modal-fade-enter-from .upload-modal,.modal-fade-leave-to .upload-modal,.track-list-enter-from,.track-list-leave-to{transform:none;filter:none}.nav-item,.upload-button,.drop-zone,.row-play,.player-art{transition:color .15s ease,background .15s ease,border-color .15s ease}}
</style>
