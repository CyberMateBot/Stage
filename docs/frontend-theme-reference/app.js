/**
 * Reference implementation (vanilla JS). Port to React/Vue in a separate frontend repo.
 * API base MUST come from env (Vite: import.meta.env.VITE_API_URL), not window.location.origin.
 */
(function () {
  const STORAGE_KEY = "cybermate-ui-theme";
  const THEMES = ["light", "dark"];

  const tg = window.Telegram && window.Telegram.WebApp;
  const statusEl = document.getElementById("status");
  const buttons = Array.from(document.querySelectorAll(".theme-btn"));

  function apiBase() {
    return (typeof import.meta !== "undefined" && import.meta.env && import.meta.env.VITE_API_URL) ||
      window.__API_BASE__ ||
      "http://localhost:8090";
  }

  function setStatus(text, isError) {
    statusEl.textContent = text || "";
    statusEl.classList.toggle("is-error", Boolean(isError));
  }

  function telegramTheme() {
    if (!tg) return null;
    return tg.colorScheme === "dark" ? "dark" : "light";
  }

  function applyTheme(theme, options) {
    const opts = options || {};
    const next = THEMES.includes(theme) ? theme : "light";
    document.documentElement.setAttribute("data-theme", next);
    localStorage.setItem(STORAGE_KEY, next);

    buttons.forEach(function (btn) {
      btn.classList.toggle("is-active", btn.dataset.themeValue === next);
    });

    if (tg) {
      try {
        if (next === "dark") {
          tg.setHeaderColor?.("#0b1220");
          tg.setBackgroundColor?.("#0b1220");
        } else {
          tg.setHeaderColor?.("#f4f6fb");
          tg.setBackgroundColor?.("#f4f6fb");
        }
      } catch (_) {}
    }

    if (!opts.silent) {
      setStatus(next === "dark" ? "Тёмная тема" : "Светлая тема");
    }
    return next;
  }

  function telegramId() {
    if (!tg?.initDataUnsafe?.user) return null;
    return String(tg.initDataUnsafe.user.id);
  }

  async function fetchServerTheme(id) {
    const res = await fetch(apiBase() + "/v1/users/telegram/" + encodeURIComponent(id));
    if (!res.ok) return null;
    const body = await res.json();
    const theme = body?.data?.theme;
    return THEMES.includes(theme) ? theme : null;
  }

  async function saveServerTheme(id, theme) {
    const res = await fetch(
      apiBase() + "/v1/users/telegram/" + encodeURIComponent(id) + "/theme",
      {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ theme }),
      }
    );
    if (!res.ok) throw new Error("save failed: " + res.status);
    return res.json();
  }

  async function resolveInitialTheme() {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (THEMES.includes(stored)) return stored;

    const id = telegramId();
    if (id) {
      try {
        const remote = await fetchServerTheme(id);
        if (remote) return remote;
      } catch (_) {}
    }

    return telegramTheme() || "light";
  }

  async function selectTheme(theme) {
    const applied = applyTheme(theme);
    const id = telegramId();
    if (!id) return;

    try {
      await saveServerTheme(id, applied);
      setStatus("Тема сохранена");
    } catch (_) {
      setStatus("Тема применена локально (сервер недоступен)", true);
    }
  }

  buttons.forEach(function (btn) {
    btn.addEventListener("click", function () {
      selectTheme(btn.dataset.themeValue);
    });
  });

  if (tg) {
    tg.ready();
    tg.expand();
    tg.onEvent("themeChanged", function () {
      if (!localStorage.getItem(STORAGE_KEY)) {
        applyTheme(telegramTheme() || "light", { silent: true });
      }
    });
  }

  resolveInitialTheme()
    .then(function (theme) {
      applyTheme(theme, { silent: true });
      setStatus("");
    })
    .catch(function () {
      applyTheme("light", { silent: true });
    });
})();
