// The package root re-exports the connect family for one-line consumers;
// precise consumers import the "./connect" subpath so future dialects
// never widen their graph.
export * from "./connect";
