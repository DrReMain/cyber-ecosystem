FROM node:24-alpine AS build
ARG NPM_REGISTRY=https://registry.npmjs.org/
WORKDIR /src
RUN corepack enable
# Install with the FULL workspace source present. Do not "optimize" this
# into a manifests-only install layer: pnpm yields a different node_modules
# shape from a manifest-only workspace, and vite's server environment then
# hashes ?url css assets differently than the client environment emits —
# phantom css filenames in the SSR bundle that 404 at runtime (first-paint
# flash). Verified in both directions. The pnpm store cache mount keeps
# reinstalls fast despite the busted layer.
COPY . .
RUN --mount=type=cache,id=pnpm-store,target=/pnpm/store \
    pnpm install --frozen-lockfile --ignore-scripts --store-dir /pnpm/store \
    --filter @cyber-ecosystem/admin... --registry="${NPM_REGISTRY}"
RUN pnpm --filter @cyber-ecosystem/admin build

FROM node:24-alpine
ARG NPM_REGISTRY=https://registry.npmjs.org/
WORKDIR /app
COPY --from=build /src/pnpm-workspace.yaml /src/pnpm-lock.yaml /src/package.json /src/.npmrc ./
COPY --from=build /src/shared-ts/antd/package.json            ./shared-ts/antd/package.json
COPY --from=build /src/shared-ts/cookie/package.json          ./shared-ts/cookie/package.json
COPY --from=build /src/shared-ts/error/package.json           ./shared-ts/error/package.json
COPY --from=build /src/shared-ts/router-progress/package.json ./shared-ts/router-progress/package.json
COPY --from=build /src/shared-ts/store/package.json           ./shared-ts/store/package.json
COPY --from=build /src/shared-ts/theme/package.json           ./shared-ts/theme/package.json
COPY --from=build /src/gen/connect-ts/package.json            ./gen/connect-ts/package.json
COPY --from=build /src/gen/openapi-ts/package.json            ./gen/openapi-ts/package.json
COPY --from=build /src/app/clients/admin/package.json         ./app/clients/admin/package.json
RUN corepack enable && pnpm install --prod --frozen-lockfile --filter @cyber-ecosystem/admin --registry="${NPM_REGISTRY}"
COPY --from=build /src/app/clients/admin/dist ./app/clients/admin/dist
COPY --from=build /src/app/clients/admin/server.mjs ./app/clients/admin/server.mjs
WORKDIR /app/app/clients/admin
ENV NODE_ENV=production PORT=3000
USER node
CMD ["node", "server.mjs"]
