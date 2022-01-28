const production = !process.env.ROLLUP_WATCH;
module.exports = {
  theme: {
    extend: {
      colors: {
        "dark-primary": "#212121",
        "green-basic": "#09D9C6",
        "purple-basic": "#C496FC",
      },
    },
    fontFamily: {
      sans: ["Fira Sans"],
      mono: ["Inconsolata"],
    },
  },
  future: {
    purgeLayersByDefault: true,
    removeDeprecatedGapUtilities: true,
  },
  plugins: [],
  purge: {
    content: ["./src/App.svelte"],
    enabled: production, // disable purge in dev
  },
};
