import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import Instances from './pages/Instances';
import InstanceDetail from './pages/InstanceDetail';
import ApiDocs from './pages/ApiDocs';

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Instances />} />
          <Route path="/instance/:name" element={<InstanceDetail />} />
          <Route path="/docs" element={<ApiDocs />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
