<template>
  <router-view />
</template>

<style>
*,
*::before,
*::after {
  box-sizing: border-box;
}

html {
  background: #0E0E0E;
}

body {
  min-width: 320px;
  min-height: 100vh;
  margin: 0;
  background: #0E0E0E;
  font-family: "Manrope", system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
}

button,
input {
  font: inherit;
}

button,
a {
  -webkit-tap-highlight-color: transparent;
}

::selection {
  background: #FFCC00;
  color: #111111;
}
</style>
