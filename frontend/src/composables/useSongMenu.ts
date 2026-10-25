import type { DropdownOption } from "naive-ui";
import type { Track } from "@/api/types";
import { useMusicStore } from "@/stores";
import { usePlayerController } from "@/core/player/PlayerController";
import { renderIcon, copyData } from "@/utils/helper";
import {
  openCopySongInfo,
  openJumpArtist,
  openPlaylistAdd,
} from "@/utils/modal";
import { deleteSongs } from "@/utils/auth";

/**
 * 歌曲菜单。照搬 SPlayer composables/useSongMenu.ts，条目按阶段 2 的“歌曲列表”：
 * 立即播放、下一首播放、添加到歌单、更多操作（复制歌曲名称、复制更多信息）、
 * 从歌单中删除（在自己的歌单中时）、查看专辑、查看艺术家、同名搜索（在曲库中搜索）。
 * SPlayer 没有“查看专辑”与“查看艺术家”，图标沿用其 Album 与 Artist。
 * 菜单项的显示设置（contextMenuOptions）不迁入。
 */
export const useSongMenu = () => {
  const router = useRouter();
  const musicStore = useMusicStore();
  const player = usePlayerController();

  // 生成菜单选项
  const getMenuOptions = (
    song: Track,
    playlistId: number = 0,
    emit?: (event: "removeSong", args: number[]) => void,
  ): DropdownOption[] => {
    const isCurrent = musicStore.playSong.id === song.id;

    return [
      {
        key: "play",
        label: "立即播放",
        props: {
          onClick: () => player.addNextSong(song, true),
        },
        icon: renderIcon("Play", { size: 18 }),
      },
      {
        key: "play-next",
        label: "下一首播放",
        show: !isCurrent,
        props: {
          onClick: () => player.addNextSong(song, false),
        },
        icon: renderIcon("PlayNext", { size: 18 }),
      },
      {
        key: "playlist-add",
        label: "添加到歌单",
        props: {
          onClick: () => openPlaylistAdd([song]),
        },
        icon: renderIcon("AddList", { size: 18 }),
      },
      {
        key: "line-1",
        type: "divider",
      },
      {
        key: "more",
        label: "更多操作",
        icon: renderIcon("Menu", { size: 18 }),
        children: [
          {
            key: "code-name",
            label: "复制歌曲名称",
            props: {
              onClick: () => copyData(song.title),
            },
            icon: renderIcon("Copy", { size: 18 }),
          },
          {
            key: "copy-song-info",
            label: "复制更多信息",
            props: {
              onClick: () => openCopySongInfo(song.id),
            },
            icon: renderIcon("FormatList", { size: 18 }),
          },
        ],
      },
      {
        key: "line-two",
        type: "divider",
      },
      {
        key: "delete-playlist",
        label: "从歌单中删除",
        show: emit !== undefined && !!playlistId,
        props: {
          // 写入在阶段 5 接入，成功后由页面从列表中移除（emit removeSong）
          onClick: () => deleteSongs(playlistId, [song.id]),
        },
        icon: renderIcon("Delete"),
      },
      {
        key: "album",
        label: "查看专辑",
        show: !!song.album,
        props: {
          onClick: () =>
            router.push({ name: "album", query: { id: song.album?.id } }),
        },
        icon: renderIcon("Album"),
      },
      {
        key: "artist",
        label: "查看艺术家",
        show: song.artists.length > 0,
        props: {
          // 一位艺术家直接跳转，多位时选择
          onClick: () =>
            openJumpArtist(
              song.artists,
              song.artists.length === 1 ? song.artists[0].id : undefined,
            ),
        },
        icon: renderIcon("Artist"),
      },
      {
        key: "search",
        label: "同名搜索",
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
