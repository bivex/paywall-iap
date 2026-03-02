/**
 * Dev-only error logger with component stack support.
 * In production this is a no-op (tree-shaken out).
 *
 * Usage:
 *   logError(error, "Dashboard::fetchItems")
 *
 * Or wrap risky code:
 *   try { ... } catch (e) { logError(e, "MyComponent"); throw e; }
 */
export function logError(error: unknown, context: string): void {
  if (process.env.NODE_ENV !== "development") return;

  const err = error instanceof Error ? error : new Error(String(error));

  console.group(
    `%c💥 ${context}`,
    "color: #dc2626; font-weight: bold; font-size: 13px"
  );
  console.error(`${err.name}: ${err.message}`);

  // Component stack (React error boundary provides this)
  const compStack = (error as { componentStack?: string })?.componentStack;
  if (compStack) {
    console.group("📍 Component tree (most precise):");
    console.log(compStack.trim());
    console.groupEnd();
  }

  // JS stack — highlight app frames
  if (err.stack) {
    const frames = err.stack
      .split("\n")
      .slice(1)
      .map((f) => f.trim());

    const appFrames = frames.filter(
      (f) =>
        /\bsrc\/|\/app\/|\/actions\/|\/components\//.test(f) &&
        !f.includes("node_modules")
    );

    if (appFrames.length > 0) {
      console.group("🔥 Your code frames:");
      appFrames.forEach((f) =>
        console.log("%c" + f, "color: #f97316; font-weight: 600")
      );
      console.groupEnd();
    }

    console.groupCollapsed("Full JS stack:");
    frames.forEach((f) => console.log(f));
    console.groupEnd();
  }

  console.groupEnd();
}

/**
 * Safe accessor — returns fallback instead of throwing on null/undefined.
 *
 * Usage:
 *   const count = safe(() => items.length, 0)
 */
export function safe<T>(fn: () => T, fallback: T): T {
  try {
    const v = fn();
    return v ?? fallback;
  } catch {
    return fallback;
  }
}
