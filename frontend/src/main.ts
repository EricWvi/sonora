import { createApp } from "vue";
import { createPinia } from "pinia";
import persistedstate from "pinia-plugin-persistedstate";
import App from "./App.vue";
import router from "./router";
import "./style/main.scss";
import "./style/animate.scss";
import {
  debounceDirective,
  throttleDirective,
  visibleDirective,
} from "@/utils/instruction";

const pinia = createPinia();
pinia.use(persistedstate);
const app = createApp(App);
app.use(pinia).use(router);
app.directive("debounce", debounceDirective);
app.directive("throttle", throttleDirective);
app.directive("visible", visibleDirective);
app.mount("#app");
