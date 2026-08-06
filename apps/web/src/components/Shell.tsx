import { NavLink, Outlet } from "react-router-dom";

const destinations = [
  ["/nearby", "Nearby"],
  ["/map", "Map"],
  ["/plan", "Plan"],
  ["/alerts", "Alerts"],
  ["/saved", "Saved"],
] as const;

export function Shell() {
  return (
    <>
      <a className="skip-link" href="#main">
        Skip to content
      </a>
      <div className="shell">
        <header className="topbar">
          <NavLink className="brand" to="/nearby">
            Tabi
          </NavLink>
          <nav aria-label="Primary">
            {destinations.map(([to, label]) => (
              <NavLink key={to} to={to}>
                {label}
              </NavLink>
            ))}
          </nav>
        </header>
        <aside className="sidebar">
          <nav aria-label="Primary">
            {destinations.map(([to, label]) => (
              <NavLink key={to} to={to}>
                {label}
              </NavLink>
            ))}
          </nav>
          <NavLink to="/settings/privacy">Settings</NavLink>
          <NavLink to="/credits">Credits</NavLink>
        </aside>
        <main id="main" tabIndex={-1}>
          <Outlet />
        </main>
      </div>
    </>
  );
}
