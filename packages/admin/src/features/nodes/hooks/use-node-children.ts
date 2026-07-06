import { useQuery } from "@tanstack/react-query";
import { useClient } from "../../../providers/client-context";

// Lists the direct children of a directory. `parentId === null` lists root via
// `listRootChildren`; otherwise it lists the children of the given directory.
export function useNodeChildren(parentId: string | null) {
  const client = useClient();
  return useQuery({
    queryKey: ["nodes", "children", parentId],
    queryFn: async () => {
      const response =
        parentId === null
          ? await client.listRootChildren({})
          : await client.listChildren({ id: parentId });
      return response.nodes;
    },
  });
}
