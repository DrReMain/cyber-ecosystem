// The provider gates devtools on import.meta.env.DEV — vite injects it in the
// consuming app. Declare the minimal shape so the package typechecks without
// depending on vite types; merges cleanly with the app's vite/client types.
interface ImportMetaEnv {
  readonly DEV: boolean;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

// tsc 7.0.2 tolerates relative css imports but errors on PACKAGE-subpath css
// even with a `declare module "*.css"` wildcard — declare the exact specifier
// (jotai-devtools exports ./styles.css → dist/index.css).
declare module "*.css";
declare module "jotai-devtools/styles.css";
