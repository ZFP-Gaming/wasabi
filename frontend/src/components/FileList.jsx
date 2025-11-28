import { useMemo, useRef, useState } from "react";
import {
  Clock,
  FileAudio,
  NotePencil,
  PauseCircle,
  PlayCircle,
  Trash,
  WaveSine,
} from "phosphor-react";
import { fileUrl } from "../services/api.js";

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

function formatTime(seconds) {
  if (!Number.isFinite(seconds)) return "0:00";
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60)
    .toString()
    .padStart(2, "0");
  return `${mins}:${secs}`;
}

function FileCard({ file, onRename, onDelete, disabled }) {
  const [editing, setEditing] = useState(false);
  const [newName, setNewName] = useState(file.name);
  const [localError, setLocalError] = useState("");
  const [playing, setPlaying] = useState(false);
  const audioSource = useMemo(() => fileUrl(file.name), [file.name]);
  const audioRef = useRef(null);

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

  const handleDelete = async () => {
    if (disabled) return;
    const confirmed = window.confirm(`¿Eliminar ${file.name}?`);
    if (!confirmed) return;
    await onDelete(file.name);
  };

  const togglePlay = () => {
    const audio = audioRef.current;
    if (!audio) return;
    if (playing) {
      audio.pause();
      setPlaying(false);
      return;
    }
    audio.play();
  };

  return (
    <article className="file-card">
      <div className="file-header compact">
        <div className="file-row">
          <div className="file-avatar">
            <FileAudio size={20} weight="fill" />
          </div>
          <div className="file-name-line">
            <p className="file-name">{file.name}</p>
            <div className="file-meta inline">
              <span className="pill subtle compact tiny-pill">
                <Clock size={14} weight="bold" />
                {formatDate(file.modified)}
              </span>
              <span className="pill subtle compact tiny-pill">{formatSize(file.size)}</span>
            </div>
          </div>
        </div>
        {!editing ? (
          <div className="chip-actions">
            <button
              className="ghost icon-button"
              onClick={() => setEditing(true)}
              disabled={disabled}
              aria-label="Renombrar"
              title="Renombrar"
            >
              <NotePencil size={16} weight="bold" />
            </button>
            <button
              className="ghost danger icon-button"
              onClick={handleDelete}
              disabled={disabled}
              aria-label="Eliminar"
              title="Eliminar"
            >
              <Trash size={16} weight="bold" />
            </button>
          </div>
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

      <div className="player-shell compact micro">
        <button
          type="button"
          className="primary-circle small"
          onClick={togglePlay}
          disabled={disabled}
          aria-label={playing ? "Pausar" : "Reproducir"}
          title={playing ? "Pausar" : "Reproducir"}
        >
          {playing ? <PauseCircle size={26} weight="fill" /> : <PlayCircle size={26} weight="fill" />}
        </button>
        <div className="player-meta horizontal micro">
          <WaveSine size={14} weight="bold" />
          <span className="muted tiny">Reproducir</span>
        </div>
        <audio
          ref={audioRef}
          className="sr-audio"
          preload="metadata"
          src={audioSource}
          controlsList="nodownload noplaybackrate"
          onPlay={() => setPlaying(true)}
          onPause={() => setPlaying(false)}
          onEnded={() => setPlaying(false)}
          onLoadedMetadata={() => audioRef.current?.duration}
        />
      </div>
    </article>
  );
}

function FileList({ files, loading, onRename, onDelete, disabled }) {
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
          onDelete={onDelete}
          disabled={disabled}
        />
      ))}
    </div>
  );
}

export default FileList;
