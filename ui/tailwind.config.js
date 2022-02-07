const production = !process.env.ROLLUP_WATCH;
const purpleBasic = "#C496FC";
const greenBasic = "#09D9C6";
const whiteBasic = "#DEE4E7";

module.exports = {
  theme: {
    extend: {
      // typography: {
      //   DEFAULT: {
      //     css: {
      //       "color": whiteBasic,
      //       "h1,h2,h3,h4": {
      //         "color": greenBasic,
      //       },
      //       "a": {
      //         "color": purpleBasic,
      //       },
      //       'code': {
      //         "color": whiteBasic
      //       },
      //       'code::before': {
      //         content: '""'
      //       },
      //       'code::after': {
      //         content: '""'
      //       }
      //     }
      //   }
      // },
      colors: {
        "dark-primary": "#212121",
        "green-basic": greenBasic,
        "purple-basic": purpleBasic,
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
  plugins: [
    // require('@tailwindcss/typography'),
  ],
  purge: {
    content: ["./src/App.svelte"],
    enabled: production, // disable purge in dev
  },
};
