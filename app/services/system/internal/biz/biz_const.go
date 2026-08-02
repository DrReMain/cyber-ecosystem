package biz

// defaultTenant is the tenant id used in single-tenant deployments (D2
// degenerate core). Multi-tenant deployments use real tenant ids; this moves
// to config when tenant becomes deployment-configurable.
const defaultTenant = "default"
