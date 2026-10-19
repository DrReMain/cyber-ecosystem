import { createErrorDomain } from "@cyber-ecosystem/shared-error";
import { describe, expect, it } from "vitest";
import { viewTransitionSource } from "./view-transition-source";

const domain = createErrorDomain({ sources: [viewTransitionSource] });

describe("viewTransitionSource", () => {
  it("classifies the hidden-document transition abort as silent canceled", () => {
    const abort = new DOMException(
      "Transition was aborted because of invalid state. Document hidden",
      "InvalidStateError",
    );
    const envelope = domain.normalize(abort);
    expect(envelope.kind).toBe("canceled");
    expect(envelope.reason).toBe("VIEW_TRANSITION_ABORTED");
    expect(domain.view(abort).silent).toBe(true);
  });

  it("prefix-matches sibling abort reasons, not only the hidden-document one", () => {
    const abort = new DOMException("Transition was aborted by user script", "AbortError");
    expect(domain.normalize(abort).kind).toBe("canceled");
  });

  it("leaves unrelated InvalidStateErrors and plain errors alone", () => {
    expect(
      viewTransitionSource.recognize(new DOMException("play() failed", "InvalidStateError")),
    ).toBeUndefined();
    expect(viewTransitionSource.recognize(new Error("Transition was NOT aborted"))).toBeUndefined();
    expect(viewTransitionSource.recognize("a bare string")).toBeUndefined();
  });
});
