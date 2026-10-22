import { createApp } from "vue";
import { createPinia } from "pinia";
import persistedstate from "pinia-plugin-persistedstate";
import App from "./App.vue";
import router from "./router";
import "./style/main.scss";
import "./style/animate.scss";

const pinia = createPinia();
pinia.use(persistedstate);
createApp(App).use(pinia).use(router).mount("#app");
