import { useEffect, useState } from "react";
import FileList from "./components/FileList.jsx";
import UploadForm from "./components/UploadForm.jsx";
import { fetchFiles, renameFile, uploadFile } from "./services/api.js";

function App() {
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");

  const loadFiles = async () => {
    setLoading(true);
    setError("");
    try {
      const data = await fetchFiles();
      setFiles(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadFiles();
  }, []);

  const handleUpload = async (formData) => {
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const res = await uploadFile(formData);
      setNotice(`Archivo subido: ${res.name}`);
      await loadFiles();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  };

  const handleRename = async (currentName, newName) => {
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const res = await renameFile(currentName, newName);
      setNotice(`Archivo renombrado a ${res.name}`);
      await loadFiles();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="page">
      <header className="hero">
        <div>
          <p className="eyebrow">Wasabi · gestor de audio</p>
          <h1>Sube, lista y renombra tus audios sin complicaciones.</h1>
          <p className="lede">
            La API en Go ya está corriendo en el puerto 8080. Este cliente en
            React consume los endpoints <code>/upload</code> y{" "}
            <code>/files</code> vía proxy para evitar CORS.
          </p>
        </div>
        <div className="pill">
          <span className="dot" />
          API en vivo · 8080
        </div>
      </header>

      <div className="alerts">
        {error && <div className="alert alert-error">{error}</div>}
        {notice && <div className="alert alert-success">{notice}</div>}
      </div>

      <section className="panel">
        <div className="panel-header">
          <div>
            <h2>Subida rápida</h2>
            <p className="muted">
              Elige un archivo de audio y opcionalmente define el nombre final.
            </p>
          </div>
          <button className="ghost" onClick={loadFiles} disabled={loading}>
            Recargar lista
          </button>
        </div>
        <UploadForm onSubmit={handleUpload} busy={busy} />
      </section>

      <section className="panel">
        <div className="panel-header">
          <div>
            <h2>Archivos en servidor</h2>
            <p className="muted">
              Consulta los archivos guardados en la carpeta <code>uploads</code>
              .
            </p>
          </div>
          {loading ? (
            <span className="pill subtle">Cargando...</span>
          ) : (
            <span className="pill subtle">
              {files.length} archivo{files.length === 1 ? "" : "s"}
            </span>
          )}
        </div>
        <FileList
          files={files}
          loading={loading}
          onRename={handleRename}
          disabled={busy}
        />
      </section>
    </div>
  );
}

export default App;
