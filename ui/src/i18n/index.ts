import i18next from "i18next";
import { initReactI18next } from "react-i18next";
import en from "./en";
import zh from "./zh";

export function getLanguage(): string {
  return navigator.language;
}

// declare module "i18next" {
//   interface CustomTypeOptions {
//     defaultNS: "translation";
//     resources: {
//       translation: TT;
//     };
//   }
// }

i18next
  .use(initReactI18next)
  .init({
    resources: {
      zh: {
        translation: zh,
      },
      en: {
        translation: en,
      },
    },
    lng: getLanguage(),
    fallbackLng: "en",
    interpolation: {
      escapeValue: false,
    },
  })
  .then();
