// 照搬 SPlayer core/automix/SharedAudioContext.ts。Sonora 没有“音频延迟模式”设置，
// latencyHint 固定为 SPlayer 的默认值 interactive。

/** 扩充 AudioContext 接口以支持 setSinkId（照搬 SPlayer types/audio/context.ts） */
export interface IExtendedAudioContext extends AudioContext {
  setSinkId(deviceId: string): Promise<void>;
}

/** 共享音频上下文 */
let sharedContext: IExtendedAudioContext | null = null;
/** 主输入节点 */
let masterInput: GainNode | null = null;
/** 主限制器节点 */
let masterLimiter: DynamicsCompressorNode | null = null;

/**
 * 获取共享音频上下文
 * @returns 共享音频上下文
 */
export const getSharedAudioContext = (): IExtendedAudioContext => {
  if (!sharedContext) {
    sharedContext = new AudioContext({
      latencyHint: "interactive",
    }) as IExtendedAudioContext;
  }
  return sharedContext;
};

/**
 * 获取主输入节点
 * @returns 主输入节点
 */
export const getSharedMasterInput = (): GainNode => {
  const ctx = getSharedAudioContext();
  if (!masterInput) {
    masterInput = ctx.createGain();
    masterLimiter = ctx.createDynamicsCompressor();

    masterLimiter.threshold.value = -1;
    masterLimiter.knee.value = 0;
    masterLimiter.ratio.value = 20;
    masterLimiter.attack.value = 0.003;
    masterLimiter.release.value = 0.25;

    masterInput.connect(masterLimiter);
    masterLimiter.connect(ctx.destination);
  }
  return masterInput;
};
