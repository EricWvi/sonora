import { h } from "vue";
import { NA, NText, type DataTableColumn } from "naive-ui";

interface Identified {
  id: number;
  neteaseId: number | null;
}

/** The library ID, the first column of the track, album and artist tables. */
export const idColumn = <T extends Identified>(): DataTableColumn<T> => ({
  title: "ID",
  key: "id",
  width: 80,
});

/** What a NetEase ID names, which picks its page on NetEase's site. */
export type NeteaseKind = "song" | "album" | "artist";

/** The page of a song, album or artist on NetEase's site. */
export const neteaseURL = (kind: NeteaseKind, id: number) =>
  `https://music.163.com/#/${kind}?id=${id}`;

/**
 * The NetEase ID an import or the admin recorded, linking to its page on
 * NetEase's site in a new tab, or a dash for uploads.
 */
export const neteaseIdColumn = <T extends Identified>(
  kind: NeteaseKind,
): DataTableColumn<T> => ({
  title: "网易云 ID",
  key: "neteaseId",
  width: 120,
  render: (row) =>
    row.neteaseId === null
      ? h(NText, { depth: 3 }, () => "—")
      : h(
          NA,
          {
            href: neteaseURL(kind, row.neteaseId),
            target: "_blank",
            rel: "noopener noreferrer",
            title: "在网易云音乐中打开",
          },
          () => String(row.neteaseId),
        ),
});

/**
 * The tables list the newest first. A search keeps its relevance order, which
 * puts the closest matches on the first page.
 */
export const newestFirst = (q: string) =>
  q ? {} : { sort: "createTime", order: "desc" };
