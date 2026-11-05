import type { DropdownOption } from "naive-ui";
import { albumRoute, isOnline, sameTrack } from "@/utils/track";
import type { Track } from "@/api/types";
import { useMusicStore, useSettingStore } from "@/stores";
import { usePlayerController } from "@/core/player/PlayerController";
import { renderIcon, copyData } from "@/utils/helper";
import { openJumpArtist, openPlaylistAdd } from "@/utils/modal";
import { deleteSongs } from "@/utils/auth";

/**
 * 歌曲菜单。照搬 SPlayer composables/useSongMenu.ts，条目按阶段 2 的“歌曲列表”：
 * 立即播放、下一首播放、添加到歌单、更多操作（复制歌曲名称）、
 * 从歌单中删除（在自己的歌单中时）、查看专辑、查看艺术家、同名搜索（在曲库中搜索）。
 * SPlayer 没有“查看专辑”与“查看艺术家”，图标沿用其 Album 与 Artist。
 * 菜单项的显示设置 contextMenuOptions 照搬，只有 Sonora 的条目；“更多操作”只剩复制歌曲名称，
 * 关掉复制歌曲名称时“更多操作”一并隐藏。
 */
export const useSongMenu = () => {
  const router = useRouter();
  const musicStore = useMusicStore();
  const settingStore = useSettingStore();
  const player = usePlayerController();

  // 生成菜单选项
  const getMenuOptions = (
    song: Track,
    playlistId: number = 0,
    emit?: (event: "removeSong", args: number[]) => void,
  ): DropdownOption[] => {
    const isCurrent = sameTrack(musicStore.playSong, song);
    const options = settingStore.contextMenuOptions;

    return [
      {
        key: "play",
        label: "立即播放",
        show: options.play,
        props: {
          onClick: () => player.addNextSong(song, true),
        },
        icon: renderIcon("Play", { size: 18 }),
      },
      {
        key: "play-next",
        label: "下一首播放",
        show: options.playNext && !isCurrent,
        props: {
          onClick: () => player.addNextSong(song, false),
        },
        icon: renderIcon("PlayNext", { size: 18 }),
      },
      {
        key: "playlist-add",
        label: "添加到歌单",
        // 本地歌单只放曲库中的歌曲，在线歌曲要先导入
        show: options.addToPlaylist && !isOnline(song),
        props: {
          onClick: () => openPlaylistAdd([song]),
        },
        icon: renderIcon("AddList", { size: 18 }),
      },
      {
        key: "line-1",
        type: "divider",
        show: options.play || options.playNext || options.addToPlaylist,
      },
      {
        key: "more",
        label: "更多操作",
        show: options.more && options.copyName,
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
        ],
      },
      {
        key: "line-two",
        type: "divider",
        show: options.more && options.copyName,
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
