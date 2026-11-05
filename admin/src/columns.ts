import { h } from "vue";
import { NText, type DataTableColumn } from "naive-ui";

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

/** The NetEase ID an import recorded, or a dash for uploads. */
export const neteaseIdColumn = <
  T extends Identified,
>(): DataTableColumn<T> => ({
  title: "网易云 ID",
  key: "neteaseId",
  width: 120,
  render: (row) => row.neteaseId ?? h(NText, { depth: 3 }, () => "—"),
});

/**
 * The tables list the newest first. A search keeps its relevance order, which
 * puts the closest matches on the first page.
 */
export const newestFirst = (q: string) =>
  q ? {} : { sort: "createTime", order: "desc" };
