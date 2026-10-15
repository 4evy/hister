export function modal(node: HTMLDialogElement) {
  let pointerStartedOutside = false;

  function isOutside(event: MouseEvent) {
    const bounds = node.getBoundingClientRect();
    return event.target === node && isOutsideBounds(event, bounds);
  }

  function isOutsideBounds(event: MouseEvent, bounds: DOMRect) {
    return (
      event.clientX < bounds.left ||
      event.clientX > bounds.right ||
      event.clientY < bounds.top ||
      event.clientY > bounds.bottom
    );
  }

  function handlePointerDown(event: PointerEvent) {
    pointerStartedOutside = isOutside(event);
  }

  function handleClick(event: MouseEvent) {
    if (pointerStartedOutside && isOutside(event)) node.close();
    pointerStartedOutside = false;
  }

  node.addEventListener('pointerdown', handlePointerDown);
  node.addEventListener('click', handleClick);
  node.showModal();

  return {
    destroy() {
      node.removeEventListener('pointerdown', handlePointerDown);
      node.removeEventListener('click', handleClick);
      node.close();
    },
  };
}
