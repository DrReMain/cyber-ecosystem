# Skin presets law card

- Files under `shared-ts/antd/src/skin-presets/**` — the law lives in `docs/conventions/skin/CONVENTIONS.md`.
- One file per skin, default-exported; registration in `SKINS` is the only wiring step.
- Every chrome token contract slot must be filled — a missing slot falls back to antd defaults somewhere in the layout.
- New-skin recipe and the Storybook tuning loop: `docs/conventions/skin/RUNBOOKS.md`.
