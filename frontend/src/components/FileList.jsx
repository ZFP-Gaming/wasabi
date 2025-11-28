import { useMemo, useState } from "react";

function formatSize(bytes) {
  if (!bytes) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  const exponent = Math.min(
    Math.floor(Math.log(bytes) / Math.log(1024)),
    units.length - 1,
  );
  const value = bytes / 1024 ** exponent;
  return `${value.toFixed(value >= 10 ? 0 : 1)} ${units[exponent]}`;
}

function formatDate(dateString) {
  const date = new Date(dateString);
  return date.toLocaleString("es-ES", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function FileCard({ file, onRename, disabled }) {
  const [editing, setEditing] = useState(false);
  const [newName, setNewName] = useState(file.name);
  const [localError, setLocalError] = useState("");

  const handleRename = async (event) => {
    event.preventDefault();
    setLocalError("");
    if (!newName.trim()) {
      setLocalError("Ingresa un nombre válido");
      return;
    }
    if (newName.trim() === file.name) {
      setEditing(false);
      return;
    }
    await onRename(file.name, newName.trim());
    setEditing(false);
  };

  return (
    <article className="file-card">
      <div className="file-header">
        <div>
          <p className="file-name">{file.name}</p>
          <p className="muted">
            {formatSize(file.size)} · {formatDate(file.modified)}
          </p>
        </div>
        {!editing ? (
          <button
            className="ghost"
            onClick={() => setEditing(true)}
            disabled={disabled}
          >
            Renombrar
          </button>
        ) : (
          <button
            className="ghost"
            onClick={() => {
              setEditing(false);
              setNewName(file.name);
              setLocalError("");
            }}
          >
            Cancelar
          </button>
        )}
      </div>

      {editing && (
        <form className="rename-form" onSubmit={handleRename}>
          <label className="field">
            <span>Nuevo nombre</span>
            <input
              type="text"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              disabled={disabled}
              required
            />
          </label>
          <div className="actions">
            <button type="submit" className="primary" disabled={disabled}>
              Guardar
            </button>
            {localError && <p className="muted error">{localError}</p>}
          </div>
        </form>
      )}
    </article>
  );
}

function FileList({ files, loading, onRename, disabled }) {
  const orderedFiles = useMemo(
    () =>
      [...files].sort(
        (a, b) => new Date(b.modified).getTime() - new Date(a.modified).getTime(),
      ),
    [files],
  );

  if (loading) {
    return (
      <div className="file-grid skeleton">
        <div className="skeleton-card" />
        <div className="skeleton-card" />
        <div className="skeleton-card" />
      </div>
    );
  }

  if (!orderedFiles.length) {
    return <p className="muted">No hay archivos en el servidor todavía.</p>;
  }

  return (
    <div className="file-grid">
      {orderedFiles.map((file) => (
        <FileCard
          key={file.name}
          file={file}
          onRename={onRename}
          disabled={disabled}
        />
      ))}
    </div>
  );
}

export default FileList;
