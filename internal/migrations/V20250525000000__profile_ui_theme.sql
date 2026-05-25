ALTER TABLE profiles
    ADD COLUMN IF NOT EXISTS ui_theme TEXT NOT NULL DEFAULT 'light'
        CHECK (ui_theme IN ('light', 'dark'));
