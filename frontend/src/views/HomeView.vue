<template>
  <div class="home">
    <div class="welcome">
      <n-h1>{{ greeting }}</n-h1>
      <n-text depth="3">由此开启好心情 ~</n-text>
    </div>
    <div class="rec-public">
      <n-flex class="title" align="center" justify="space-between">
        <n-h3 prefix="bar"><n-text>推荐歌单</n-text></n-h3>
      </n-flex>
      <n-empty description="空空如也，怎么什么都没有啊" size="large" />
    </div>
  </div>
</template>

<script setup lang="ts">
// 沿用参考页面问候语，展示初始化阶段不读取在线账号。
const hour = new Date().getHours();
const greetings: [number, string][] = [
  [6, "凌晨好"],
  [9, "早上好"],
  [12, "上午好"],
  [14, "中午好"],
  [17, "下午好"],
  [19, "傍晚好"],
  [22, "晚上好"],
  [24, "夜深了"],
];
const greeting = greetings.find(([limit]) => hour < limit)?.[1] ?? "夜深了";
</script>

<style lang="scss" scoped>
.home {
  width: 100%;
  max-width: 1500px;
  margin: 0 auto;
  .welcome {
    margin-top: 8px;
    margin-bottom: 20px;
    .n-h1 {
      margin: 0;
      font-weight: bold;
    }
  }
}
.title {
  margin-top: 28px;
  padding: 0 4px;
  width: max-content;
  .n-h {
    margin: 0;
    display: flex;
    align-items: center;
  }
}
</style>
