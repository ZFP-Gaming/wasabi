export function splitFileName(name = "") {
  const trimmed = name.trim();
  const lastDot = trimmed.lastIndexOf(".");
  if (lastDot <= 0) {
    return { base: trimmed, ext: "" };
  }
  return {
    base: trimmed.slice(0, lastDot),
    ext: trimmed.slice(lastDot),
  };
}

export function displayName(name = "") {
  const { base } = splitFileName(name);
  return base || name;
}

export function ensureExtension(name = "", originalName = "") {
  const trimmed = name.trim();
  if (!trimmed) return trimmed;

  const hasExtension = /\.[^./]+$/.test(trimmed);
  if (hasExtension) return trimmed;

  const { ext } = splitFileName(originalName);
  if (!ext) return trimmed;
  return trimmed.endsWith(ext) ? trimmed : `${trimmed}${ext}`;
}
