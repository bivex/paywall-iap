import { describe, it, expect } from "vitest";
import {
  parsePaywallDefinition,
  stringifyPaywallDefinition,
  DEFAULT_PAYWALL_TEMPLATE,
} from "@/lib/paywall-schema";

describe("parsePaywallDefinition", () => {
  it("parses valid paywall JSON", () => {
    const json = JSON.stringify(DEFAULT_PAYWALL_TEMPLATE);
    const result = parsePaywallDefinition(json);
    expect(result.success).toBe(true);
  });

  it("returns error for empty string", () => {
    const result = parsePaywallDefinition("");
    expect(result.success).toBe(false);
  });

  it("returns error for plain text (e.g. '404 page not found')", () => {
    const result = parsePaywallDefinition("404 page not found");
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.errors.length).toBeGreaterThan(0);
    }
  });

  it("returns error for double-encoded JSON string", () => {
    // Backend bug: definition serialized twice → string instead of object
    const doubleEncoded = JSON.stringify(JSON.stringify(DEFAULT_PAYWALL_TEMPLATE));
    const result = parsePaywallDefinition(doubleEncoded);
    expect(result.success).toBe(false);
  });

  it("returns error for invalid JSON", () => {
    const result = parsePaywallDefinition("{invalid json}");
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.errors[0]).toMatch(/unexpected|invalid|expected|json/i);
    }
  });

  it("returns error for missing required fields", () => {
    const result = parsePaywallDefinition(JSON.stringify({ version: "1" }));
    expect(result.success).toBe(false);
  });
});

describe("stringifyPaywallDefinition", () => {
  it("produces valid JSON that re-parses correctly", () => {
    const str = stringifyPaywallDefinition(DEFAULT_PAYWALL_TEMPLATE);
    expect(() => JSON.parse(str)).not.toThrow();
    const result = parsePaywallDefinition(str);
    expect(result.success).toBe(true);
  });
});

describe("loadPaywall simulation (definition from API response)", () => {
  it("handles definition as object (normal case)", () => {
    // Backend returns definition as parsed JSON object
    const definitionFromApi = DEFAULT_PAYWALL_TEMPLATE as unknown as Record<string, unknown>;
    const schemaText = JSON.stringify(definitionFromApi, null, 2);
    const result = parsePaywallDefinition(schemaText);
    expect(result.success).toBe(true);
  });

  it("handles definition as string (double-encoded backend bug)", () => {
    // If backend double-encodes: definition is a string, not object
    const definitionFromApi = JSON.stringify(DEFAULT_PAYWALL_TEMPLATE);
    // JSON.stringify on a string produces quoted string
    const schemaText = JSON.stringify(definitionFromApi, null, 2);
    // This will look like '"{ ... }"' — invalid for paywall schema
    const result = parsePaywallDefinition(schemaText);
    // Should fail gracefully, not throw SyntaxError
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.errors.length).toBeGreaterThan(0);
    }
  });

  it("handles null definition gracefully", () => {
    const schemaText = JSON.stringify(null, null, 2);
    const result = parsePaywallDefinition(schemaText);
    expect(result.success).toBe(false);
  });
});
