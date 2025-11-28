import { useEffect, useRef, useState } from "react";
import { useAuth } from "../contexts/AuthContext";
import { CaretDown, SignOut } from "phosphor-react";

function UserProfile() {
  const { user, logout } = useAuth();
  const [open, setOpen] = useState(false);
  const menuRef = useRef(null);

  if (!user) return null;

  const getAvatarUrl = () => {
    if (!user.avatar) {
      const defaultAvatarIndex = parseInt(user.discriminator || "0") % 5;
      return `https://cdn.discordapp.com/embed/avatars/${defaultAvatarIndex}.png`;
    }
    return `https://cdn.discordapp.com/avatars/${user.user_id}/${user.avatar}.png`;
  };

  const displayName = user.discriminator && user.discriminator !== "0"
    ? `${user.username}#${user.discriminator}`
    : user.username;

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (menuRef.current && !menuRef.current.contains(event.target)) {
        setOpen(false);
      }
    };

    const handleEscape = (event) => {
      if (event.key === "Escape") {
        setOpen(false);
      }
    };

    document.addEventListener("click", handleClickOutside);
    document.addEventListener("keydown", handleEscape);

    return () => {
      document.removeEventListener("click", handleClickOutside);
      document.removeEventListener("keydown", handleEscape);
    };
  }, []);

  return (
    <div className={`user-profile ${open ? "is-open" : ""}`} ref={menuRef}>
      <button className="user-info" onClick={() => setOpen(!open)} type="button">
        <div className="user-avatar-frame">
          <img
            src={getAvatarUrl()}
            alt={user.username}
            className="user-avatar"
          />
        </div>
        <div className="user-details">
          <span className="user-name">{displayName}</span>
          <span className="user-badge">Discord</span>
        </div>
        <CaretDown size={16} weight="bold" className="user-caret" />
      </button>
      {open && (
        <div className="user-menu">
          <button className="menu-item" onClick={logout} type="button">
            <SignOut size={16} weight="bold" />
            Salir
          </button>
        </div>
      )}
    </div>
  );
}

export default UserProfile;
