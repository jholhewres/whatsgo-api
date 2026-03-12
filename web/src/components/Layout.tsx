import { Link, Outlet, useLocation } from 'react-router-dom';

const navItems = [
  { path: '/', label: 'Instances' },
  { path: '/docs', label: 'API Docs' },
];

export default function Layout() {
  const location = useLocation();

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="sidebar-header">
          <h1>WhatsGo</h1>
          <span className="version">v1.0.0</span>
        </div>
        <nav>
          {navItems.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              className={`nav-item ${location.pathname === item.path ? 'active' : ''}`}
            >
              {item.label}
            </Link>
          ))}
        </nav>
      </aside>
      <main className="content">
        <Outlet />
      </main>
    </div>
  );
}
