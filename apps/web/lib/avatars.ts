export const AVATARS = [
  { id: 1, src: "/assets/avatar/avatar1.png", label: "Avatar 1" },
  { id: 2, src: "/assets/avatar/avatar2.png", label: "Avatar 2" },
  { id: 3, src: "/assets/avatar/avatar3.png", label: "Avatar 3" },
  { id: 4, src: "/assets/avatar/avatar4.png", label: "Avatar 4" },
  { id: 5, src: "/assets/avatar/avatar5.png", label: "Avatar 5" },
  { id: 6, src: "/assets/avatar/avatar6.png", label: "Avatar 6" },
  { id: 7, src: "/assets/avatar/avatar7.png", label: "Avatar 7" },
  { id: 8, src: "/assets/avatar/avatar8.png", label: "Avatar 8" },
];

export function getAvatarSrc(id: number): string {
  const avatar = AVATARS.find((a) => a.id === id);
  return avatar ? avatar.src : AVATARS[0].src;
}
