import { toast } from "sonner";

export const fileUpload = async ({
  event,
  onProgress,
  onSuccess,
}: {
  event: React.ChangeEvent<HTMLInputElement>;
  onProgress: (progress: number) => void;
  onSuccess: (response: string) => void;
}) => {
  const files = event.target.files;
  if (!files) return;

  const xhr = new XMLHttpRequest();
  const formData = new FormData();
  formData.append("photos", files[0], files[0].name);

  xhr.open("POST", "/api/upload");

  xhr.upload.onprogress = (event) => {
    if (event.lengthComputable) {
      const progress = Math.round((event.loaded / event.total) * 100);
      onProgress(progress);
    }
  };

  xhr.onload = () => {
    if (xhr.status >= 200 && xhr.status < 300) {
      onSuccess(xhr.responseText);
    } else {
      toast.error("Upload Failed");
    }
  };

  xhr.onerror = () => toast.error("Network error");

  xhr.send(formData);
};
