import { defineStore } from 'pinia'
import i18n from '../i18n'

export const useAppStore = defineStore('app', {
  state: () => {
    const savedLang = localStorage.getItem('app_language')
    let defaultLang = 'en-US'
    
    if (savedLang) {
      defaultLang = savedLang
    } else {
      const navLang = navigator.language
      if (navLang.startsWith('zh')) {
        defaultLang = 'zh-CN'
      }
    }

    // 同步初始化 i18n 实例的语言
    if (i18n.mode === 'legacy') {
      (i18n.global.locale as any) = defaultLang
    } else {
      (i18n.global.locale as any).value = defaultLang
    }

    return {
      language: defaultLang
    }
  },
  actions: {
    setLanguage(lang: string) {
      this.language = lang
      localStorage.setItem('app_language', lang)
      
      // 动态更新 vue-i18n
      if (i18n.mode === 'legacy') {
        (i18n.global.locale as any) = lang
      } else {
        (i18n.global.locale as any).value = lang
      }
    }
  }
})
