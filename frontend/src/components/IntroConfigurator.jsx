import { useMemo, useRef, useState } from "react";
import {
  CheckCircle,
  Headphones,
  MagnifyingGlass,
  MusicNotesSimple,
  PauseCircle,
  PlayCircle,
} from "phosphor-react";
import { fileUrl } from "../services/api.js";
import { displayName } from "../utils/fileNames.js";

function IntroConfigurator({ files, loading, busy, onSubmit }) {
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState("");
  const [localError, setLocalError] = useState("");
   const [showSuggestions, setShowSuggestions] = useState(false);
   const [highlightedIndex, setHighlightedIndex] = useState(-1);
  const [playing, setPlaying] = useState(false);
  const audioRef = useRef(null);

  const orderedFiles = useMemo(
    () =>
      [...files].sort(
        (a, b) => new Date(b.modified).getTime() - new Date(a.modified).getTime(),
      ),
    [files],
  );

  const resolveSelectionFromQuery = (value) => {
    const term = value.trim().toLowerCase();
    if (!term) return "";
    return (
      orderedFiles.find((file) => displayName(file.name).toLowerCase() === term)?.name ||
      ""
    );
  };

  const filtered = useMemo(() => {
    const term = query.trim().toLowerCase();
    if (!term) return orderedFiles.slice(0, 8);
    return orderedFiles
      .filter((file) => displayName(file.name).toLowerCase().includes(term))
      .slice(0, 12);
  }, [orderedFiles, query]);

  const suggestions = filtered;

  const selectSound = (name) => {
    if (audioRef.current) {
      audioRef.current.pause();
      audioRef.current.currentTime = 0;
      setPlaying(false);
    }
    setSelected(name);
    setQuery(displayName(name));
    setLocalError("");
    setShowSuggestions(false);
    setHighlightedIndex(-1);
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    const existingName = selected || resolveSelectionFromQuery(query);
    const trimmed = existingName.trim();
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

  const handleKeyDown = (event) => {
    if (!suggestions.length) return;
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setShowSuggestions(true);
      setHighlightedIndex((prev) => (prev + 1) % suggestions.length);
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      setShowSuggestions(true);
      setHighlightedIndex((prev) => (prev - 1 + suggestions.length) % suggestions.length);
    } else if (event.key === "Enter" && highlightedIndex >= 0) {
      event.preventDefault();
      const picked = suggestions[highlightedIndex];
      selectSound(picked.name);
    } else if (event.key === "Escape") {
      setShowSuggestions(false);
      setHighlightedIndex(-1);
    }
  };

  return (
    <div className="panel">
      <div className="panel-header">
        <div>
          <h2>Configurar mi intro</h2>
        </div>
      </div>

      <form className="intro-form" onSubmit={handleSubmit}>
        <label className="field">
          <span>Elige un sonido de la carpeta uploads</span>
          <div className="autocomplete">
            <div className="input-shell">
              <MagnifyingGlass size={18} weight="bold" className="muted" />
              <input
                type="text"
                placeholder="escribe para filtrar..."
                value={query}
                onFocus={() => setShowSuggestions(true)}
                onBlur={() => setTimeout(() => setShowSuggestions(false), 120)}
                onChange={(e) => {
                  const value = e.target.value;
                  setQuery(value);
                  setLocalError("");
                  setShowSuggestions(true);
                  setHighlightedIndex(-1);
                  const match = resolveSelectionFromQuery(value);
                  if (match) {
                    setSelected(match);
                  } else {
                    setSelected("");
                  }
                }}
                onKeyDown={handleKeyDown}
                disabled={busy || loading}
              />
              {selected && (
                <CheckCircle size={16} weight="fill" className="input-shell__status" />
              )}
            </div>
            {showSuggestions && suggestions.length > 0 && (
              <div className="autocomplete__list">
                {suggestions.map((file, index) => {
                  const isHighlighted = index === highlightedIndex;
                  return (
                    <button
                      key={file.name}
                      type="button"
                      className={`autocomplete__item ${isHighlighted ? "is-highlighted" : ""}`}
                      onMouseDown={(event) => event.preventDefault()}
                      onClick={() => selectSound(file.name)}
                    >
                      <span className="autocomplete__name" title={file.name}>
                        {displayName(file.name)}
                      </span>
                      <span className="autocomplete__meta">
                        Editado {new Date(file.modified).toLocaleDateString("es-ES")}
                      </span>
                    </button>
                  );
                })}
              </div>
            )}
          </div>
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
                  <strong className="sound-name" title={file.name}>
                    {displayName(file.name)}
                  </strong>
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
              <span className="muted tiny" title={selected}>
                {displayName(selected)}
              </span>
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
          Guardar
        </button>
      </form>
    </div>
  );
}

export default IntroConfigurator;
