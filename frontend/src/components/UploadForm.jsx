import { useState } from "react";

function UploadForm({ onSubmit, busy }) {
  const [file, setFile] = useState(null);
  const [customName, setCustomName] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (event) => {
    event.preventDefault();
    if (!file) {
      return;
    }

    const formData = new FormData();
    formData.append("file", file);
    if (customName.trim()) {
      formData.append("filename", customName.trim());
    }

    setSubmitting(true);
    try {
      await onSubmit(formData);
      setFile(null);
      setCustomName("");
      event.target.reset();
    } finally {
      setSubmitting(false);
    }
  };

  const disabled = busy || submitting;

  return (
    <form className="upload-form" onSubmit={handleSubmit}>
      <label className="field">
        <span>Archivo</span>
        <input
          type="file"
          accept="audio/*"
          onChange={(e) => setFile(e.target.files?.[0] ?? null)}
          disabled={disabled}
          required
        />
      </label>

      <label className="field">
        <span>Nombre opcional</span>
        <input
          type="text"
          placeholder="mi-audio.wav"
          value={customName}
          onChange={(e) => setCustomName(e.target.value)}
          disabled={disabled}
        />
      </label>

      <button type="submit" className="primary" disabled={disabled}>
        {submitting ? "Subiendo..." : "Subir archivo"}
      </button>
    </form>
  );
}

export default UploadForm;
