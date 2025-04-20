import { i18n, ThemeProvider } from "@allape/gocrud-react";
import { Locale } from "antd/es/locale";
import zhCN from "antd/locale/zh_CN";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.scss";
import App from "./App.tsx";
import en from "./i18n/en.ts";
import zh from "./i18n/zh.ts";

function getLocale(): Locale | undefined {
  const language = i18n.getLanguage();
  if (language.startsWith("zh")) {
    import("dayjs/locale/zh-cn");
    return zhCN;
  }
  return undefined;
}

i18n
  .setup({
    zh,
    en,
  })
  .then(() => {
    createRoot(document.getElementById("root")!).render(
      <StrictMode>
        <ThemeProvider locale={getLocale()}>
          <App />
        </ThemeProvider>
      </StrictMode>,
    );
  });
