import { useCallback, useEffect, useState } from "react";
import { Compass, Sliders } from "phosphor-react";
import { useAuth } from "./contexts/AuthContext";
import Login from "./components/Login";
import UserProfile from "./components/UserProfile";
import FileList from "./components/FileList.jsx";
import UploadForm from "./components/UploadForm.jsx";
import IntroConfigurator from "./components/IntroConfigurator.jsx";
import { deleteFile, fetchFiles, renameFile, sendIntroRequest, uploadFile } from "./services/api.js";

function App() {
  const { user, loading: authLoading } = useAuth();
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");
  const [activeView, setActiveView] = useState("sounds");

  const loadFiles = useCallback(async () => {
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
  }, []);

  useEffect(() => {
    if (!user) {
      setFiles([]);
      setLoading(false);
      return;
    }
    loadFiles();
  }, [user, loadFiles]);

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

  const handleDelete = async (name) => {
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const res = await deleteFile(name);
      setNotice(`Archivo eliminado: ${res.name}`);
      await loadFiles();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  };

  const handleIntroRequest = async (soundName) => {
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const res = await sendIntroRequest(soundName);
      setNotice(`Solicitud de intro registrada con ${res.soundName}`);
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  };

  if (authLoading) {
    return (
      <div className="page loading-page">
        <div className="loader">Cargando...</div>
      </div>
    );
  }

  if (!user) {
    return <Login />;
  }

  return (
    <div className="page">
      <header className="hero">
        <div>
          <p className="eyebrow">Wasabi · gestor de audio</p>
        </div>
        <div className="hero-actions">
          <UserProfile />
        </div>
      </header>

      <div className="app-shell">
        <aside className="sidebar">
          <nav className="nav">
            <button
              className={`nav-item ${activeView === "sounds" ? "is-active" : ""}`}
              onClick={() => setActiveView("sounds")}
              type="button"
            >
              <Sliders size={16} weight="bold" />
              Administrar sonidos
            </button>
            <button
              className={`nav-item ${activeView === "intro" ? "is-active" : ""}`}
              onClick={() => setActiveView("intro")}
              type="button"
            >
              <Compass size={16} weight="bold" />
              Configurar mi intro
            </button>
          </nav>
        </aside>

        <div className="content-area">
          <div className="alerts">
            {error && <div className="alert alert-error">{error}</div>}
            {notice && <div className="alert alert-success">{notice}</div>}
          </div>

          {activeView === "intro" ? (
            <IntroConfigurator
              files={files}
              loading={loading}
              busy={busy}
              onSubmit={handleIntroRequest}
            />
          ) : (
            <>
              <section className="panel">
                <div className="panel-header">
                  <div>
                    <h2>Subida rápida</h2>
                    <p className="muted">
                      Elige un archivo de audio y opcionalmente define el nombre final.
                    </p>
                  </div>
                </div>
                <UploadForm onSubmit={handleUpload} busy={busy} />
              </section>

              <section className="panel">
                <div className="panel-header">
                  <div>
                    <h2>Archivos en servidor</h2>
                    <p className="muted">
                      Consulta los archivos guardados en la carpeta <code>uploads</code>.
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
                  onDelete={handleDelete}
                  disabled={busy}
                />
              </section>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

export default App;
