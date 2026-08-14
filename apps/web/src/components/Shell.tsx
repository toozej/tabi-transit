import { NavLink, Outlet } from "react-router-dom";

const destinations = [
  ["/nearby", "Nearby", "01"],
  ["/map", "Map", "02"],
  ["/plan", "Plan", "03"],
  ["/alerts", "Alerts", "04"],
  ["/saved", "Saved", "05"],
] as const;

export function Shell() {
  return (
    <>
      <a className="skip-link" href="#main">
        Skip to content
      </a>
      <div className="shell">
        <header className="topbar">
          <NavLink className="brand" to="/nearby" aria-label="Tabi home">
            <span aria-hidden="true" className="brand-mark">
              T
            </span>
            <span>Tabi</span>
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
          <NavLink className="brand" to="/nearby" aria-label="Tabi home">
            <span aria-hidden="true" className="brand-mark">
              T
            </span>
            <span>Tabi</span>
          </NavLink>
          <p className="sidebar-kicker">Portland transit, at hand</p>
          <nav aria-label="Primary">
            {destinations.map(([to, label, stop]) => (
              <NavLink key={to} to={to}>
                <span aria-hidden="true" className="nav-stop">
                  {stop}
                </span>
                {label}
              </NavLink>
            ))}
          </nav>
          <div className="sidebar-utility">
            <NavLink to="/settings/privacy">Settings</NavLink>
            <NavLink to="/credits">Credits</NavLink>
          </div>
        </aside>
        <main id="main" tabIndex={-1}>
          <Outlet />
        </main>
      </div>
    </>
  );
}
