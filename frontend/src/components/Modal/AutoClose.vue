<!-- 照搬 SPlayer components/Modal/AutoClose.vue。与 SPlayer 不同：
关闭开关时同时停止计时器（SPlayer 下一秒会暂停播放）；第一次打开开关时用默认时长开始
（SPlayer 的剩余时长是 0，开关打开了但没有计时）；超过 60 分钟时显示小时（SPlayer 的
mm:ss 会回绕，120 分钟显示为 00:00）。 -->
<template>
  <n-flex :size="20" class="auto-close" align="center" vertical>
    <n-card class="open">
      <n-flex align="center" justify="space-between">
        <n-flex size="small" align="center">
          <SvgIcon name="TimeAuto" size="22" />
          <Transition name="fade" mode="out-in">
            <n-text v-if="!statusStore.autoClose.enable"> 未开启 </n-text>
            <n-text v-else strong>
              {{ msToTime(statusStore.autoClose.remainTime * 1000) }}
            </n-text>
          </Transition>
        </n-flex>
        <n-switch
          v-model:value="statusStore.autoClose.enable"
          :round="false"
          @update:value="handleUpdate"
        />
      </n-flex>
    </n-card>
    <!-- 时间选择 -->
    <n-flex size="large" align="center" justify="center">
      <n-tag
        v-for="(item, index) in [10, 20, 30, 45, 60, 90, 120]"
        :bordered="false"
        :key="index"
        type="primary"
        size="large"
        round
        @click="player.startAutoCloseTimer(item, item * 60)"
      >
        {{ item }}min
      </n-tag>
      <!-- 自定义 -->
      <n-popconfirm
        :negative-text="null"
        :positive-button-props="{
          strong: true,
          secondary: true,
          type: 'primary',
        }"
        :show-icon="false"
        @positive-click="
          player.startAutoCloseTimer(customTime, customTime * 60)
        "
      >
        <template #trigger>
          <n-tag :bordered="false" type="primary" size="large" round>
            自定义时长
          </n-tag>
        </template>
        <n-flex vertical>
          <n-text>自定义时长（分钟）</n-text>
          <n-input-number
            v-model:value="customTime"
            :min="1"
            :max="120"
            placeholder="请输入自定义时长"
          />
        </n-flex>
      </n-popconfirm>
    </n-flex>
    <!-- 是否播放完 -->
    <n-checkbox v-model:checked="statusStore.autoClose.waitSongEnd">
      等待整首歌曲播放完成再停止播放
    </n-checkbox>
  </n-flex>
</template>

<script setup lang="ts">
import { useStatusStore } from "@/stores";
import { msToTime } from "@/utils/time";
import { usePlayerController } from "@/core/player/PlayerController";

const player = usePlayerController();
const statusStore = useStatusStore();

// 自定义时长
const customTime = ref(1);

// 是否开启
const handleUpdate = (value: boolean) => {
  if (value) {
    const { time, remainTime } = statusStore.autoClose;
    player.startAutoCloseTimer(time, remainTime || time * 60);
  } else {
    player.stopAutoCloseTimer();
  }
};
</script>

<style scoped lang="scss">
.auto-close {
  width: 100%;
  .open {
    width: 100%;
    border-radius: 8px;
    .n-text {
      font-size: 18px;
    }
  }
}
</style>
