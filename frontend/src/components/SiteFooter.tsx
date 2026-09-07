import { Mail, MapPin, Phone } from "lucide-react";
import { FaFacebook, FaInstagram, FaLinkedin } from "react-icons/fa";
import { CLINIC } from "../lib/clinic";
import { Logo } from "./SiteHeader";

const SOCIALS = [
  { href: CLINIC.social.instagram, label: "Instagram", Icon: FaInstagram },
  { href: CLINIC.social.facebook, label: "Facebook", Icon: FaFacebook },
  { href: CLINIC.social.linkedin, label: "LinkedIn", Icon: FaLinkedin },
];

export function SiteFooter() {
  return (
    <footer className="mt-auto bg-navy-deep px-5 py-14 text-navy-mist">
      <div className="mx-auto grid max-w-4xl gap-10 md:grid-cols-[1.3fr_1fr]">
        <div>
          <div className="mb-4 flex items-center gap-3">
            <Logo size={40} />
            <span>
              <span className="block font-display text-lg leading-tight text-white">
                {CLINIC.name}
              </span>
              <span className="block font-deva text-sm text-navy-mist">{CLINIC.nameDeva}</span>
            </span>
          </div>
          <p className="max-w-sm text-sm leading-relaxed">
            A family hospital and diagnostics centre in {CLINIC.city}, open every hour of every
            day. Consultation fees are payable at the clinic.
          </p>
          <p className="mt-5 flex items-start gap-2 text-sm">
            <MapPin size={16} className="mt-0.5 shrink-0 text-red" aria-hidden />
            {CLINIC.address}
          </p>
        </div>

        <div className="md:justify-self-end">
          <h2 className="mb-4 text-sm font-medium tracking-wide text-white">Reach us</h2>
          <ul className="space-y-3 text-sm">
            <li>
              <a
                href={CLINIC.phoneHref}
                className="inline-flex items-center gap-2 transition-colors hover:text-white"
              >
                <Phone size={16} aria-hidden /> {CLINIC.phoneDisplay}
              </a>
            </li>
            <li>
              <a
                href={`mailto:${CLINIC.email}`}
                className="inline-flex items-center gap-2 break-all transition-colors hover:text-white"
              >
                <Mail size={16} aria-hidden /> {CLINIC.email}
              </a>
            </li>
          </ul>

          <ul className="mt-5 flex gap-4">
            {SOCIALS.map(({ href, label, Icon }) => (
              <li key={label}>
                <a
                  href={href}
                  target="_blank"
                  rel="noreferrer"
                  aria-label={label}
                  className="inline-block transition-colors hover:text-white"
                >
                  <Icon size={20} aria-hidden />
                </a>
              </li>
            ))}
          </ul>
        </div>
      </div>

      <p className="mx-auto mt-12 max-w-6xl border-t border-navy-soft pt-6 text-center text-xs">
        © {new Date().getFullYear()} {CLINIC.nameFull}. All rights reserved.
      </p>
    </footer>
  );
}
