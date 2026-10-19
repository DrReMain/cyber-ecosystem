// css side-effect imports (view-transition.css, route-transition.css) have
// no type meaning — declare them so any TS version resolves the module.
// consuming apps usually get this from vite/client; a shared package must
// not depend on vite types, so keep the minimal declaration here.
declare module "*.css";
