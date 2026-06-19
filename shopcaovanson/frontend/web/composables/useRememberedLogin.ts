import {
  getRememberPreference,
  getSavedPassword,
  listSavedEmails,
  loadLastRememberedCredentials,
  removeRememberedAccount,
  saveRememberedAccount,
  setRememberPreference,
} from '~/utils/rememberedLogin'

export const useRememberedLogin = () => {
  const savedEmails = ref<string[]>([])
  const rememberMe = ref(true)
  const ready = ref(false)

  const refreshSavedEmails = () => {
    if (!import.meta.client) {
      return
    }
    savedEmails.value = listSavedEmails()
  }

  const init = async () => {
    if (!import.meta.client) {
      return null
    }
    rememberMe.value = getRememberPreference()
    refreshSavedEmails()
    ready.value = true
    return loadLastRememberedCredentials()
  }

  const persistAfterLogin = async (email: string, password: string) => {
    if (!import.meta.client) {
      return
    }
    setRememberPreference(rememberMe.value)
    if (rememberMe.value) {
      await saveRememberedAccount(email, password)
      refreshSavedEmails()
    }
  }

  const fillPasswordForEmail = async (email: string) => {
    if (!email.trim()) {
      return null
    }
    return getSavedPassword(email)
  }

  const forgetAccount = (email: string) => {
    removeRememberedAccount(email)
    refreshSavedEmails()
  }

  return {
    savedEmails,
    rememberMe,
    ready,
    init,
    persistAfterLogin,
    fillPasswordForEmail,
    forgetAccount,
    refreshSavedEmails,
  }
}
