(function () {
  const STORAGE_KEY = "cybermate-ui-theme";
  const THEMES = ["light", "dark"];
  const API = window.__API_BASE__ || "http://localhost:8090";
  const tg = window.Telegram?.WebApp;
  const buttons = document.querySelectorAll(".theme-btn");

  function applyTheme(theme) {
    const next = THEMES.includes(theme) ? theme : "light";
    document.documentElement.setAttribute("data-theme", next);
    localStorage.setItem(STORAGE_KEY, next);
    buttons.forEach((b) => b.classList.toggle("is-active", b.dataset.themeValue === next));
    if (tg) {
      const color = next === "dark" ? "#0b1220" : "#f4f6fb";
      tg.setHeaderColor?.(color);
      tg.setBackgroundColor?.(color);
    }
    return next;
  }

  async function fetchTheme(id) {
    const r = await fetch(`${API}/v1/users/telegram/${encodeURIComponent(id)}`);
    if (!r.ok) return null;
    const j = await r.json();
    return THEMES.includes(j?.data?.theme) ? j.data.theme : null;
  }

  async function saveTheme(id, theme) {
    await fetch(`${API}/v1/users/telegram/${encodeURIComponent(id)}/theme`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ theme }),
    });
  }

  buttons.forEach((btn) => btn.addEventListener("click", async () => {
    const t = applyTheme(btn.dataset.themeValue);
    const id = tg?.initDataUnsafe?.user?.id && String(tg.initDataUnsafe.user.id);
    if (id) try { await saveTheme(id, t); } catch (_) {}
  }));

  (async () => {
    const stored = localStorage.getItem(STORAGE_KEY);
    let theme = THEMES.includes(stored) ? stored : null;
    const id = tg?.initDataUnsafe?.user?.id && String(tg.initDataUnsafe.user.id);
    if (!theme && id) theme = await fetchTheme(id).catch(() => null);
    if (!theme) theme = tg?.colorScheme === "dark" ? "dark" : "light";
    applyTheme(theme);
    tg?.ready();
    tg?.expand();
  })();
})();
