import type { DropdownOption } from "naive-ui";
import {
  albumRoute,
  canImportSong,
  isOnline,
  isPlayable,
  sameTrack,
} from "@/utils/track";
import type { Track } from "@/api/types";
import { useMusicStore, useSettingStore, useUserStore } from "@/stores";
import { usePlayerController } from "@/core/player/PlayerController";
import { renderIcon, copyData } from "@/utils/helper";
import {
  openCopySongInfo,
  openJumpArtist,
  openPlaylistAdd,
} from "@/utils/modal";
import { copyShareLink } from "@/utils/onlineLink";
import { deleteSongs, importSong } from "@/utils/auth";
import { fillLyrics } from "@/utils/onlineFill";

/**
 * 歌曲菜单。照搬 SPlayer composables/useSongMenu.ts，条目按阶段 2 的“歌曲列表”：
 * 立即播放、下一首播放、添加到歌单、更多操作（复制歌曲名称）、
 * 从歌单中删除（在自己的歌单中时）、查看专辑、查看艺术家、同名搜索（在曲库中搜索）。
 * SPlayer 没有“查看专辑”与“查看艺术家”，图标沿用其 Album 与 Artist。
 * 菜单项的显示设置 contextMenuOptions 照搬，只有 Sonora 的条目；“更多操作”只剩复制歌曲名称，
 * 关掉复制歌曲名称时“更多操作”一并隐藏，除非还有“从网易云补全歌词”。
 * 与 SPlayer 不同（Eric，2026-11-07）：有导入权限的用户在关联了网易云、没有歌词的本地歌曲上
 * 多一项“从网易云补全歌词”。
 */
export const useSongMenu = () => {
  const router = useRouter();
  const musicStore = useMusicStore();
  const settingStore = useSettingStore();
  const player = usePlayerController();
  const userStore = useUserStore();

  // 生成菜单选项
  const getMenuOptions = (
    song: Track,
    playlistId: number = 0,
    emit?: (event: "removeSong", args: number[]) => void,
  ): DropdownOption[] => {
    const isCurrent = sameTrack(musicStore.playSong, song);
    const options = settingStore.contextMenuOptions;
    const showPlay = options.play && isPlayable(song);
    const showNext = options.playNext && !isCurrent && isPlayable(song);
    // 本地歌单只放曲库中的歌曲；在线歌曲只能加入网易云歌单，所以没有登录时不显示
    const showAdd =
      options.addToPlaylist && (!isOnline(song) || userStore.onlineSignedIn);
    const showFillLyrics =
      !isOnline(song) &&
      !!userStore.user?.canImport &&
      !!song.onlineId &&
      !song.hasLyrics;
    const showMore = options.more && (options.copyName || showFillLyrics);

    return [
      {
        key: "play",
        label: "立即播放",
        show: showPlay,
        props: {
          onClick: () => player.addNextSong(song, true),
        },
        icon: renderIcon("Play", { size: 18 }),
      },
      {
        key: "play-next",
        label: "下一首播放",
        show: showNext,
        props: {
          onClick: () => player.addNextSong(song, false),
        },
        icon: renderIcon("PlayNext", { size: 18 }),
      },
      {
        key: "playlist-add",
        label: "添加到歌单",
        show: showAdd,
        props: {
          onClick: () => openPlaylistAdd([song]),
        },
        icon: renderIcon("AddList", { size: 18 }),
      },
      {
        key: "import",
        label: "导入到曲库",
        show: canImportSong(song, userStore.user),
        props: {
          onClick: () => importSong(song),
        },
        icon: renderIcon("Download", { size: 18 }),
      },
      {
        key: "line-1",
        type: "divider",
        show:
          showPlay ||
          showNext ||
          showAdd ||
          canImportSong(song, userStore.user),
      },
      {
        key: "more",
        label: "更多操作",
        show: showMore,
        icon: renderIcon("Menu", { size: 18 }),
        children: [
          {
            key: "code-name",
            label: "复制歌曲名称",
            show: options.copyName,
            props: {
              onClick: () => copyData(song.title),
            },
            icon: renderIcon("Copy", { size: 18 }),
          },
          {
            key: "fill-lyrics",
            label: "从网易云补全歌词",
            show: showFillLyrics,
            props: {
              onClick: () => fillLyrics(song),
            },
            icon: renderIcon("Lyrics", { size: 18 }),
          },
          // 在线歌曲按 SPlayer 补齐（Eric，2026-11-06）；本地歌曲的 ID 对听众没有用，保持现状
          {
            key: "code-id",
            label: "复制歌曲 ID",
            show: isOnline(song),
            props: {
              onClick: () => copyData(song.id),
            },
            icon: renderIcon("Copy", { size: 18 }),
          },
          {
            key: "copy-song-info",
            label: "复制更多信息",
            show: isOnline(song),
            props: {
              onClick: () => openCopySongInfo(song),
            },
            icon: renderIcon("FormatList", { size: 18 }),
          },
          {
            key: "share",
            label: "分享歌曲链接",
            show: isOnline(song),
            props: {
              onClick: () =>
                copyShareLink("song", song.id, "已复制分享链接到剪切板"),
            },
            icon: renderIcon("Share", { size: 18 }),
          },
        ],
      },
      {
        key: "line-two",
        type: "divider",
        show: showMore,
      },
      {
        key: "delete-playlist",
        label: "从歌单中删除",
        show: options.deleteFromPlaylist && emit !== undefined && !!playlistId,
        props: {
          onClick: () =>
            deleteSongs(playlistId, [song.id], {
              callback: () => emit?.("removeSong", [song.id]),
              songName: song.title,
            }),
        },
        icon: renderIcon("Delete"),
      },
      {
        key: "album",
        label: "查看专辑",
        show: options.album && !!song.album,
        props: {
          onClick: () => router.push(albumRoute(song)),
        },
        icon: renderIcon("Album"),
      },
      {
        key: "artist",
        label: "查看艺术家",
        show: options.artist && song.artists.length > 0,
        props: {
          // 一位艺术家直接跳转，多位时选择
          onClick: () =>
            openJumpArtist(
              song.artists,
              // 在线歌曲没有多位艺术家的选择弹窗，进入第一位的页面
              song.artists.length === 1 || isOnline(song)
                ? song.artists[0].id
                : undefined,
              isOnline(song),
            ),
        },
        icon: renderIcon("Artist"),
      },
      {
        key: "search",
        label: "同名搜索",
        show: options.search,
        props: {
          onClick: () =>
            router.push({ name: "search", query: { keyword: song.title } }),
        },
        icon: renderIcon("Search"),
      },
    ];
  };

  return { getMenuOptions };
};
