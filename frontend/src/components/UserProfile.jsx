import { useAuth } from "../contexts/AuthContext";
import { SignOut, User } from "phosphor-react";

function UserProfile() {
  const { user, logout } = useAuth();

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

  return (
    <div className="user-profile">
      <div className="user-info">
        <img
          src={getAvatarUrl()}
          alt={user.username}
          className="user-avatar"
        />
        <div className="user-details">
          <span className="user-name">{displayName}</span>
          <span className="user-badge">Discord</span>
        </div>
      </div>
      <button className="ghost" onClick={logout} title="Cerrar sesión">
        <SignOut size={18} weight="bold" />
        Salir
      </button>
    </div>
  );
}

export default UserProfile;
