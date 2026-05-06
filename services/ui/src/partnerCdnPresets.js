const HERO_IMAGE = 'https://cdn.shopify.com/s/files/1/0258/7102/4181/files/animal-hero-image-white.jpg?v=1718127777'

const STORE_IMAGES = [
  {
    title: 'International Elephant Foundation',
    image_url: 'https://cdn.shopify.com/s/files/1/0258/7102/4181/files/international-elephant-foundation.jpg?v=1717692395',
    link_url: 'https://animalpride.com/ief',
  },
  {
    title: 'Mayor Max',
    image_url: 'https://cdn.shopify.com/s/files/1/0258/7102/4181/files/mayor-max.jpg?v=1717692521',
    link_url: 'https://animalpride.com/mm',
  },
  {
    title: 'Labrador Retriever',
    image_url: 'https://cdn.shopify.com/s/files/1/0258/7102/4181/files/labrador-retriever.jpg?v=1717692491',
    link_url: 'https://animalpride.com/lrc',
  },
  {
    title: 'Jacksonville Humane Society',
    image_url: 'https://cdn.shopify.com/s/files/1/0258/7102/4181/files/jacksonville-humane-society.jpg?v=1717692372',
    link_url: 'https://animalpride.com/jhs',
  },
  {
    title: 'Golden Retriever Rescue',
    image_url: 'https://cdn.shopify.com/s/files/1/0258/7102/4181/files/golden-retriever.jpg?v=1717692383',
    link_url: 'https://animalpride.com/grrmf',
  },
  {
    title: 'New Zoo and Adventure Park',
    image_url: 'https://cdn.shopify.com/s/files/1/0258/7102/4181/files/new-zoo-and-adventure-park.jpg?v=1717692480',
    link_url: 'https://animalpride.com/nzaap',
  },
]

const PRESETS = {
  'partnership-overview': [
    {
      type: 'image',
      heading: 'Partner With Animal Pride',
      body: "Pet and Wildlife Rescues, Clubs, Zoos and other animal-related organizations. We'll build an online store to sell your merchandise and you'll earn revenue for your organization.",
      image_url: HERO_IMAGE,
      image_alt: 'Partner with Animal Pride Hero',
      button_label: 'Apply',
      button_link: '/apply',
      background: '#00698f',
    },
    {
      type: 'buttons',
      heading: 'Explore Details',
      buttons: [
        { label: 'How It Works', url: '/how-it-works', variant: 'primary' },
        { label: 'Additional Perks', url: '/case-studies', variant: 'default' },
      ],
      background: '',
    },
  ],
  'how-it-works': [
    {
      type: 'bullets',
      heading: 'How It Works',
      items: [
        'We build your online store, help with designs/logos, and set up products at no cost.',
        'We pay your organization a percentage of every item sold through your store.',
        'You can add products from our catalog to increase revenue opportunities.',
        'You promote your store; we handle operations and fulfillment.',
      ],
      background: '#f6f7fe',
    },
    {
      type: 'form_cta',
      heading: 'Ready to Get Started?',
      body: 'Apply now and we will contact you with the next steps for onboarding.',
      button_label: 'Apply',
      button_link: '/apply',
      background: '',
    },
  ],
  'case-studies': [
    {
      type: 'bullets',
      heading: 'Additional Perks',
      items: [
        'We align your store styling with your website where possible.',
        'We print your designs across a wide variety of product types and colors.',
        'We provide marketing graphics for your web, social, and email channels.',
        'Preferred pricing is available for bulk and event-based merchandise orders.',
      ],
      background: '',
    },
    {
      type: 'text',
      heading: 'No Financial Commitment',
      body: 'You can end the relationship at any time for any reason or no reason at all.',
      background: '#f6f7fe',
    },
  ],
  'pricing-revenue-share': [
    {
      type: 'text',
      heading: 'Revenue Share',
      body: 'We structure partnerships around shared success and transparent payout expectations.',
      background: '',
    },
    {
      type: 'buttons',
      heading: 'Discuss Pricing',
      buttons: [{ label: 'Apply', url: '/apply', variant: 'primary' }],
      background: '#f6f7fe',
    },
  ],
  'partner-faq': [
    {
      type: 'image_grid',
      heading: 'Check Out Some of Our Other Stores',
      items: STORE_IMAGES,
      background: '',
    },
  ],
  'application-contact': [
    {
      type: 'text',
      heading: 'Partner Registration',
      body: 'Fill out the form and we will get back to you as soon as possible.',
      background: '',
    },
    {
      type: 'application_form',
      heading: 'Application Fields',
      submit_label: 'Submit',
      fields: [
        { name: 'contact_name', label: 'First and Last Name', type: 'text', required: true, placeholder: '' },
        { name: 'organization_name', label: 'Organization Name', type: 'text', required: true, placeholder: '' },
        { name: 'email', label: 'Email', type: 'email', required: true, placeholder: '' },
        { name: 'phone', label: 'Phone Number', type: 'tel', required: true, placeholder: '' },
        { name: 'address_line1', label: 'Address Line 1', type: 'text', required: true, placeholder: '', col: 24 },
        { name: 'address_line2', label: 'Address Line 2', type: 'text', required: false, placeholder: '', col: 24 },
        { name: 'country', label: 'Country', type: 'text', required: true, placeholder: '' },
        { name: 'city_state', label: 'City / State', type: 'text', required: true, placeholder: 'Start typing a city...', col: 24 },
        { name: 'postal_code', label: 'Postal Code', type: 'text', required: true, placeholder: '' },
        { name: 'notes', label: 'Organization Social Media Accounts', type: 'textarea', required: false, placeholder: '' },
        { name: 'goals', label: 'Message', type: 'textarea', required: false, placeholder: '' },
      ],
      background: '#f6f7fe',
    },
  ],
}

export const PARTNER_PRESET_SLUGS = Object.keys(PRESETS)

export function getPartnerPresetSections(slug) {
  const preset = PRESETS[slug] || []
  return preset.map((section) => JSON.parse(JSON.stringify(section)))
}
