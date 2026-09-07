import { lazy, Suspense } from "react";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { SiteHeader } from "./components/SiteHeader";
import { SiteFooter } from "./components/SiteFooter";
import { Spinner } from "./components/ui";
import { ErrorBoundary } from "./components/ErrorBoundary";
import { BookingWizard } from "./features/booking/BookingWizard";

// The admin panel is only ever opened by staff, so it is kept out of the
// bundle every patient downloads.
const Admin = lazy(() => import("./features/admin/Admin"));

function PublicSite() {
  return (
    <div className="flex min-h-screen flex-col">
      <SiteHeader />
      <div className="flex-1">
        <BookingWizard />
      </div>
      <SiteFooter />
    </div>
  );
}

export default function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <Suspense fallback={<Spinner label="Loading" />}>
          <Routes>
            <Route path="/admin/*" element={<Admin />} />
            <Route path="/*" element={<PublicSite />} />
          </Routes>
        </Suspense>
      </BrowserRouter>
    </ErrorBoundary>
  );
}
