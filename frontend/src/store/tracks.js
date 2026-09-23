import { defineStore } from 'pinia'
import api from '@/api.js'

const getErrorMessage = (error, fallback) => {
  return error.response?.data?.error || fallback
}

export const useTrackStore = defineStore('tracks', {
  state: () => ({
    tracks: [],
    loading: false,
    uploading: false,
    uploadProgress: 0,
    error: '',
  }),

  actions: {
    async fetchTracks() {
      this.loading = true
      this.error = ''

      try {
        const { data } = await api.get('/tracks/create')
        this.tracks = Array.isArray(data.tracks) ? data.tracks : []
      } catch (error) {
        this.error = getErrorMessage(error, 'Не удалось загрузить библиотеку.')
        throw error
      } finally {
        this.loading = false
      }
    },

    async uploadTrack(file, cover) {
      const formData = new FormData()
      formData.append('file', file)
      if (cover) formData.append('cover', cover)

      this.uploading = true
      this.uploadProgress = 0
      this.error = ''

      try {
        const { data } = await api.post('/tracks/', formData, {
          onUploadProgress: event => {
            if (event.total) {
              this.uploadProgress = Math.round((event.loaded / event.total) * 100)
            }
          },
        })

        this.tracks.unshift(data.track)
        this.uploadProgress = 100
        return data.track
      } catch (error) {
        this.error = getErrorMessage(error, 'Не удалось загрузить трек.')
        throw error
      } finally {
        this.uploading = false
      }
    },

    async getStreamUrl(trackId) {
      try {
        const { data } = await api.get(`/tracks/${trackId}/stream-url`)
        return data.url
      } catch (error) {
        this.error = getErrorMessage(error, 'Не удалось запустить трек.')
        throw error
      }
    },

    clearError() {
      this.error = ''
    },
  },
})
