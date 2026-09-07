import { useRef } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { GALLERY } from "../../lib/clinic";
import { Button, Eyebrow } from "../../components/ui";

/**
 * A scroll-snapped rail rather than a full-bleed carousel: the photographs are
 * supporting evidence that the hospital is real and staffed, so they get one
 * band of the page and let the visitor move at their own pace. Native
 * horizontal scrolling keeps it usable by touch, trackpad and keyboard; the
 * arrows are an affordance on top, not the only way through.
 */
export function PhotoRail() {
  const railRef = useRef<HTMLUListElement>(null);

  const scrollBy = (direction: 1 | -1) => {
    const rail = railRef.current;
    if (!rail) return;
    rail.scrollBy({ left: direction * rail.clientWidth * 0.8, behavior: "smooth" });
  };

  return (
    <section className="border-y border-line bg-paper-sunk/40 py-14">
      <div className="mx-auto max-w-6xl px-5">
        <div className="mb-6 flex items-end justify-between gap-4">
          <div>
            <Eyebrow en="Inside the hospital" hi="अस्पताल के भीतर" />
            <h2 className="mt-2 text-2xl md:text-3xl">Where you will be seen</h2>
          </div>
          <div className="hidden gap-2 sm:flex">
            <Button variant="ghost" onClick={() => scrollBy(-1)} aria-label="Scroll photographs left" className="px-3">
              <ChevronLeft size={16} aria-hidden />
            </Button>
            <Button variant="ghost" onClick={() => scrollBy(1)} aria-label="Scroll photographs right" className="px-3">
              <ChevronRight size={16} aria-hidden />
            </Button>
          </div>
        </div>
      </div>

      <ul
        ref={railRef}
        tabIndex={0}
        aria-label="Photographs of the hospital"
        className="mx-auto flex max-w-6xl snap-x snap-mandatory gap-4 overflow-x-auto px-5 pb-4"
      >
        {GALLERY.map((photo) => (
          <li
            key={photo.src}
            className="w-[80vw] shrink-0 snap-start sm:w-[46%] lg:w-[32%]"
          >
            <figure className="overflow-hidden rounded-card border border-line bg-white">
              <img
                src={photo.src}
                alt={photo.caption}
                loading="lazy"
                className="aspect-[4/3] w-full object-cover"
              />
              <figcaption className="px-4 py-3">
                <p className="text-sm font-medium text-navy">{photo.title}</p>
                <p className="mt-0.5 text-xs leading-relaxed text-ink-mute">{photo.caption}</p>
              </figcaption>
            </figure>
          </li>
        ))}
      </ul>
    </section>
  );
}
