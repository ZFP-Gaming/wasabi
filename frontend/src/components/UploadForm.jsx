import { useState } from "react";
import {
  CheckCircle,
  CloudArrowUp,
  MusicNotesSimple,
  Scissors,
  UploadSimple,
  WaveSine,
} from "phosphor-react";
import Mirt from "react-mirt";
import { trimAudioFile } from "../services/audio.js";
import "react-mirt/dist/css/react-mirt.css";

const ensureWavName = (name) => `${name.replace(/\.[^./]+$/, "")}.wav`;

const normalizeTime = (value, fallback = 0, max = Infinity) => {
  if (!Number.isFinite(value)) return Math.min(Math.max(0, fallback), max);
  const seconds = value > 1000 ? value / 1000 : value;
  return Math.min(Math.max(0, seconds), max);
};

const formatTime = (seconds, max = Infinity) => {
  const safe = normalizeTime(seconds, 0, max);
  const mins = Math.floor(safe / 60);
  const secs = Math.floor(safe % 60)
    .toString()
    .padStart(2, "0");
  return `${mins}:${secs}`;
};

function UploadForm({ onSubmit, busy }) {
  const [file, setFile] = useState(null);
  const [customName, setCustomName] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [processing, setProcessing] = useState(false);
  const [trimRange, setTrimRange] = useState({ start: 0, end: 0 });
  const [duration, setDuration] = useState(0);
  const [waveformLoading, setWaveformLoading] = useState(false);
  const [localError, setLocalError] = useState("");

  const resetForm = (target) => {
    setFile(null);
    setCustomName("");
    setTrimRange({ start: 0, end: 0 });
    setDuration(0);
    setWaveformLoading(false);
    setLocalError("");
    target?.reset();
  };

  const handleFileChange = (event) => {
    const selected = event.target.files?.[0] ?? null;
    setFile(selected);
    setTrimRange({ start: 0, end: 0 });
    setDuration(0);
    setLocalError("");
    setWaveformLoading(Boolean(selected));
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    if (!file) {
      return;
    }

    setLocalError("");
    setSubmitting(true);
    setProcessing(true);
    try {
      const max = duration || Infinity;
      let start = normalizeTime(trimRange.start, 0, max);
      let end = normalizeTime(trimRange.end, duration || start, max);
      if (end <= start + 0.05) {
        end = duration || end;
      }
      if (end <= start) {
        throw new Error("El recorte es demasiado corto.");
      }

      const desiredName = customName.trim();
      const trimmedFile = await trimAudioFile(
        file,
        start,
        end,
        desiredName || file.name,
      );

      const uploadName = ensureWavName(desiredName || trimmedFile.name);
      const formData = new FormData();
      formData.append("file", trimmedFile);
      formData.append("filename", uploadName);

      await onSubmit(formData);
      resetForm(event.target);
    } catch (err) {
      setLocalError(err?.message || "No se pudo recortar ni subir el archivo.");
    } finally {
      setSubmitting(false);
      setProcessing(false);
    }
  };

  const disabled = busy || submitting || processing;

  return (
    <form className="upload-form" onSubmit={handleSubmit}>
      <div className="field">
        <span>Archivo</span>
        <label
          htmlFor="file-upload"
          className={`file-drop ${file ? "has-file" : ""} ${disabled ? "is-disabled" : ""}`}
        >
          <div className="file-drop__icon">
            <CloudArrowUp size={28} weight="fill" />
          </div>
          <div className="file-drop__copy">
            <p className="file-drop__title">
              {file ? file.name : "Arrastra o selecciona un audio"}
            </p>
            <p className="muted tiny">MP3, WAV o M4A · se recorta localmente y se envía en WAV</p>
          </div>
          <div className="file-drop__cta">Elegir archivo</div>
        </label>
        <input
          id="file-upload"
          className="file-input"
          type="file"
          accept="audio/*"
          onChange={handleFileChange}
          disabled={disabled}
          required
        />
        {file && (
          <div className="file-badge">
            <MusicNotesSimple size={16} weight="fill" />
            <span>{file.type || "Audio detectado"}</span>
          </div>
        )}
      </div>

      <label className="field">
        <span>Nombre opcional</span>
        <div className="input-shell">
          <WaveSine size={18} weight="bold" className="muted" />
          <input
            type="text"
            placeholder="mi-audio.wav"
            value={customName}
            onChange={(e) => setCustomName(e.target.value)}
            disabled={disabled}
          />
          {customName && (
            <CheckCircle size={16} weight="fill" className="input-shell__status" />
          )}
        </div>
      </label>

      {file && (
        <div className="trim-panel">
          <div className="trim-panel__header">
            <div className="trim-label">
              <Scissors size={16} weight="bold" />
              <p className="muted">Recorte antes de subir</p>
            </div>
            <span className="pill subtle">
              {formatTime(trimRange.start, duration || Infinity)} ·{" "}
              {formatTime(trimRange.end || duration, duration || Infinity)}
            </span>
          </div>
          <Mirt
            file={file}
            options={{ waveformLoading }}
            onChange={({ start, end }) =>
              setTrimRange((prev) => {
                const max = duration || Infinity;
                const safeStart = normalizeTime(start, prev.start, max);
                let safeEnd = normalizeTime(end, duration || prev.end || safeStart, max);
                if (safeEnd <= safeStart) safeEnd = safeStart;
                return { start: safeStart, end: safeEnd };
              })
            }
            onAudioLoaded={(audio) => {
              const len = audio?.duration || 0;
              setDuration(len);
              setTrimRange({ start: 0, end: len });
            }}
            onWaveformLoaded={() => setWaveformLoading(false)}
            onError={(err) => setLocalError(err.message)}
          />
          <p className="muted tiny">
            El recorte ocurre en tu navegador y se envía en formato WAV conservando el nombre elegido.
          </p>
        </div>
      )}

      {localError && <p className="muted error">{localError}</p>}

      <button type="submit" className="primary" disabled={disabled}>
        {processing ? (
          "Recortando y subiendo..."
        ) : (
          <>
            <UploadSimple size={18} weight="bold" />
            Subir archivo recortado
          </>
        )}
      </button>
    </form>
  );
}

export default UploadForm;
