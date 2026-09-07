/**
 * Facts about the hospital that appear in more than one place. Kept here so a
 * changed phone number is a one-line change rather than a search.
 */
export const CLINIC = {
  name: "Tryaksh Hospital",
  nameFull: "Tryaksh Hospital & Diagnostics",
  nameDeva: "त्र्यक्ष अस्पताल",
  city: "Darbhanga",
  address: "Road No 3, Laxmisagar, Darbhanga, Bihar",
  phoneDisplay: "922 9333 922",
  phoneHref: "tel:+919229333922",
  whatsappHref: "https://wa.me/919229333922",
  email: "shankar.kr.mishra@gmail.com",
  logoUrl:
    "https://res.cloudinary.com/w5nagizy/image/upload/f_auto,q_auto/Tryaksh_1",
  timezone: "Asia/Kolkata",
  social: {
    instagram: "https://www.instagram.com/tryakshhospitalanddiagnostics/",
    facebook: "https://www.facebook.com/share/1BTSAq8Qoh/?mibextid=wwXIfr",
    linkedin:
      "https://www.linkedin.com/in/dr-shankar-mishra-md-684771160?utm_source=share_via&utm_content=profile&utm_medium=member_ios",
  },
} as const;

/** Photographs of the hospital, shown in the ward rail on the home page. */
export const GALLERY = [
  {
    src: "https://res.cloudinary.com/w5nagizy/image/upload/v1788730017/Dr.Shankar_OPD.jpg",
    title: "Outpatient department",
    caption: "Dr. Shankar Mishra seeing patients in the OPD.",
  },
  {
    src: "https://res.cloudinary.com/w5nagizy/image/upload/v1788730017/Dr.MahimaMishra_Operating.jpg",
    title: "Operating theatre",
    caption: "Dr. Mahima Mishra performing a procedure.",
  },
  {
    src: "https://res.cloudinary.com/w5nagizy/image/upload/v1788730018/Dr_Mahima_OPD.jpg",
    title: "Consultation",
    caption: "Unhurried appointments, one patient at a time.",
  },
  {
    src: "https://res.cloudinary.com/w5nagizy/image/upload/v1788730017/Dr.Mahima_Picwithpatient.jpg",
    title: "Follow-up care",
    caption: "Families who have been coming here for years.",
  },
] as const;

/**
 * The three resident doctors. Their credentials are written in Devanagari
 * because that is how they are given, and how most patients here read them.
 */
export const RESIDENT_DOCTORS = [
  {
    photo:
      "https://res.cloudinary.com/w5nagizy/image/upload/v1788728660/Dr.PashupatiMishra.png",
    name: "डॉ. पशुपति मिश्रा",
    nameLatin: "Dr. Pashupati Mishra",
    role: "वरिष्ठ सलाहकार चिकित्सक",
    roleLatin: "Senior consultant physician",
    credentials: ["एमबीबीएस (डीएमसीएच)", "डी सी पी (डीएमसीएच)"],
  },
  {
    photo:
      "https://res.cloudinary.com/w5nagizy/image/upload/v1788728756/Dr.ShankarMishra.png",
    name: "डॉ. शंकर मिश्रा, एम.डी.",
    nameLatin: "Dr. Shankar Mishra, MD",
    role: "जनरल फिजिशियन एवं पैथोलॉजिस्ट",
    roleLatin: "General physician & pathologist",
    credentials: ["एमबीबीएस (बीजेएमसी अहमदाबाद)", "एम.डी. (पीएमसीएच पटना)"],
  },
  {
    photo:
      "https://res.cloudinary.com/w5nagizy/image/upload/v1788728744/Dr.MahimaMishra.png",
    name: "डॉ. (श्रीमती) महिमा मिश्रा, एम.एस.",
    nameLatin: "Dr. (Mrs.) Mahima Mishra, MS",
    role: "स्त्री एवं प्रसूति रोग विशेषज्ञ",
    roleLatin: "Obstetrician & gynaecologist",
    credentials: ["एमबीबीएस (पीएमसीएच पटना)", "एम.एस. (पीएमसीएच पटना)"],
  },
] as const;
