import { createI18n } from 'vue-i18n'
import enUS from './locales/en-US.json'
import zhCN from './locales/zh-CN.json'

const i18n = createI18n({
  legacy: false,
  locale: 'zh-CN', // 默认语言
  fallbackLocale: 'en-US',
  messages: {
    'en-US': enUS,
    'zh-CN': zhCN
  }
})

export default i18n
