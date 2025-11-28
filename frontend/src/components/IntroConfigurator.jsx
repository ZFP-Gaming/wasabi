import { useMemo, useRef, useState } from "react";
import {
  CheckCircle,
  Compass,
  Headphones,
  MagnifyingGlass,
  MusicNotesSimple,
  PauseCircle,
  PlayCircle,
} from "phosphor-react";
import { fileUrl } from "../services/api.js";

function IntroConfigurator({ files, loading, busy, onSubmit }) {
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState("");
  const [localError, setLocalError] = useState("");
  const [playing, setPlaying] = useState(false);
  const audioRef = useRef(null);

  const orderedFiles = useMemo(
    () =>
      [...files].sort(
        (a, b) => new Date(b.modified).getTime() - new Date(a.modified).getTime(),
      ),
    [files],
  );

  const filtered = useMemo(() => {
    const term = query.trim().toLowerCase();
    if (!term) return orderedFiles.slice(0, 8);
    return orderedFiles.filter((file) => file.name.toLowerCase().includes(term)).slice(0, 12);
  }, [orderedFiles, query]);

  const selectSound = (name) => {
    if (audioRef.current) {
      audioRef.current.pause();
      audioRef.current.currentTime = 0;
      setPlaying(false);
    }
    setSelected(name);
    setQuery(name);
    setLocalError("");
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    const trimmed = selected.trim();
    if (!trimmed) {
      setLocalError("Elige un sonido válido de la lista.");
      return;
    }
    const exists = orderedFiles.some((file) => file.name === trimmed);
    if (!exists) {
      setLocalError("El sonido seleccionado no está en la lista.");
      return;
    }
    setLocalError("");
    await onSubmit(trimmed);
  };

  const togglePreview = () => {
    const audio = audioRef.current;
    if (!audio || !selected) return;
    if (playing) {
      audio.pause();
      setPlaying(false);
      return;
    }
    audio.play();
  };

  return (
    <div className="panel">
      <div className="panel-header">
        <div>
          <h2>Configurar mi intro</h2>
          <p className="muted">
            Busca entre los sonidos existentes y envía tu selección para que quede registrada.
          </p>
        </div>
        <span className="pill subtle">
          <Compass size={16} weight="bold" />
          Paso único
        </span>
      </div>

      <form className="intro-form" onSubmit={handleSubmit}>
        <label className="field">
          <span>Elige un sonido de la carpeta uploads</span>
          <div className="input-shell">
            <MagnifyingGlass size={18} weight="bold" className="muted" />
            <input
              type="text"
              placeholder="escribe para filtrar..."
              list="sound-options"
              value={query}
              onChange={(e) => {
                setQuery(e.target.value);
                setLocalError("");
              }}
              onBlur={(e) => setSelected(e.target.value)}
              disabled={busy || loading}
            />
            {selected && <CheckCircle size={16} weight="fill" className="input-shell__status" />}
          </div>
          <datalist id="sound-options">
            {orderedFiles.map((file) => (
              <option key={file.name} value={file.name} />
            ))}
          </datalist>
        </label>

        <div className="sound-grid">
          {loading ? (
            <p className="muted">Cargando sonidos...</p>
          ) : filtered.length === 0 ? (
            <p className="muted">No se encontraron sonidos que coincidan.</p>
          ) : (
            filtered.map((file) => (
              <button
                key={file.name}
                type="button"
                className={`sound-chip ${selected === file.name ? "is-active" : ""}`}
                onClick={() => selectSound(file.name)}
                disabled={busy}
              >
                <div className="sound-chip__icon">
                  <MusicNotesSimple size={16} weight="fill" />
                </div>
                <div className="sound-chip__body">
                  <strong>{file.name}</strong>
                  <span className="muted tiny">Editado {new Date(file.modified).toLocaleDateString("es-ES")}</span>
                </div>
                {selected === file.name && <CheckCircle size={16} weight="fill" className="accent" />}
              </button>
            ))
          )}
        </div>

        {selected && (
          <div className="player-shell single-line">
            <div className="player-label">
              <Headphones size={16} weight="bold" />
              <span>Previsualizar</span>
            </div>
            <div className="player-meta horizontal">
              <span className="muted tiny">{selected}</span>
            </div>
            <button
              type="button"
              className="primary-circle small"
              onClick={togglePreview}
              disabled={busy}
            >
              {playing ? <PauseCircle size={22} weight="fill" /> : <PlayCircle size={22} weight="fill" />}
            </button>
            <audio
              ref={audioRef}
              className="sr-audio"
              src={fileUrl(selected)}
              onPlay={() => setPlaying(true)}
              onPause={() => setPlaying(false)}
              onEnded={() => setPlaying(false)}
            />
          </div>
        )}

        {localError && <p className="muted error">{localError}</p>}

        <button type="submit" className="primary" disabled={busy || loading}>
          Enviar solicitud de intro
        </button>
      </form>
    </div>
  );
}

export default IntroConfigurator;
