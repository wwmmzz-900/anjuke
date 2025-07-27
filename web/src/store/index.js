import { createStore } from 'vuex'

export default createStore({
  state: () => ({
    // 文件上传相关状态
    uploadProgress: {},
    fileList: []
  }),
  getters: {
    getUploadProgress: (state) => state?.uploadProgress || {},
    getFileList: (state) => state?.fileList || []
  },
  mutations: {
    SET_UPLOAD_PROGRESS(state, { fileName, progress }) {
      state.uploadProgress[fileName] = progress
    },
    CLEAR_UPLOAD_PROGRESS(state, fileName) {
      delete state.uploadProgress[fileName]
    },
    SET_FILE_LIST(state, fileList) {
      state.fileList = fileList
    },
    ADD_FILE(state, file) {
      state.fileList.unshift(file)
    },
    REMOVE_FILE(state, fileId) {
      state.fileList = state.fileList.filter(file => file.id !== fileId)
    }
  },
  actions: {
    updateUploadProgress({ commit }, payload) {
      commit('SET_UPLOAD_PROGRESS', payload)
    },
    clearUploadProgress({ commit }, fileName) {
      commit('CLEAR_UPLOAD_PROGRESS', fileName)
    },
    setFileList({ commit }, fileList) {
      commit('SET_FILE_LIST', fileList)
    },
    addFile({ commit }, file) {
      commit('ADD_FILE', file)
    },
    removeFile({ commit }, fileId) {
      commit('REMOVE_FILE', fileId)
    }
  }
})