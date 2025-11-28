const API_BASE = import.meta.env.VITE_API_BASE || "http://localhost:8080";

async function handleResponse(response) {
  const contentType = response.headers.get("content-type");
  const payload =
    contentType && contentType.includes("application/json")
      ? await response.json()
      : null;

  if (!response.ok) {
    const message =
      (payload && (payload.message || payload.error)) ||
      `Error ${response.status}`;
    throw new Error(message);
  }

  return payload;
}

const fetchWithCredentials = (url, options = {}) => {
  return fetch(url, {
    ...options,
    credentials: "include",
  });
};

export async function fetchFiles() {
  const response = await fetchWithCredentials(`${API_BASE}/files`);
  const payload = await handleResponse(response);
  return Array.isArray(payload) ? payload : [];
}

export function fileUrl(name) {
  return `${API_BASE}/files/${encodeURIComponent(name)}`;
}

export async function uploadFile(formData) {
  const response = await fetchWithCredentials(`${API_BASE}/upload`, {
    method: "POST",
    body: formData,
  });
  return handleResponse(response);
}

export async function renameFile(currentName, newName) {
  const response = await fetchWithCredentials(`${API_BASE}/files/${encodeURIComponent(currentName)}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ newName }),
  });
  return handleResponse(response);
}

export async function deleteFile(name) {
  const response = await fetchWithCredentials(`${API_BASE}/files/${encodeURIComponent(name)}`, {
    method: "DELETE",
  });
  return handleResponse(response);
}
