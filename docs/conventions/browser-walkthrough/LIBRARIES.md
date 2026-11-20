# Browser Walkthrough — Library Operations

Operations keyed by the surface's component library, identified before
interacting (CONVENTIONS §1). The craft law lives in the CONVENTIONS beside
this file; a library is an observable surface fact (package.json, imports),
never a skeleton assumption — a project on another library adds its own
section here, nothing else changes. Judgement form throughout: phenomenon —
mechanism; prescription.

## antd

- **rc-select takes real keys.** Synthetic keyboard events are ignored:
  with the dropdown open, `press_key` Enter selects the first option,
  ↓+Enter the second — or `.click()` the option node in its portal.
- **A `Form.Item` with `name` binds exactly one controlled child.** A
  second child (explanatory text) silently unbinds the field — the Select
  looks selected, validateFields still reports empty (the
  selected-but-invalid family). Text goes outside the Item, or into its
  `extra` slot.
- **The form store outlives the modal.** `destroyOnHidden` unmounts
  fields, not values, and `initialValues` bind only on first mount — a
  reopened create-modal carries the previous entry (the
  reopen-with-old-values family). The fix is a keyed remount; the walk
  reports the family, not flaky input.
- **Wire enums are string names; option values are generated numbers** —
  an options-array reverse lookup never hits; map labels through a
  string-keyed table.
- **Overlays mount at document body.** Modal, Select, Drawer and message
  portals live outside the trigger's tree — DOM assertions and coordinate
  reasoning look at body level, not at the trigger's position.
- **Every overlay's trigger child must be a DOM node.** rc-trigger clones
  the trigger event onto the direct child — a component child (Popconfirm
  wrapping a Tooltip) drops the outer trigger's click silently: a dead
  nested trigger is an app bug family, not an input failure.

## react-aria-components (RAC)

- **A closed overlay can stay mounted.** The exit-animation state machine
  can deadlock with isOpen=false not unmounting — an invisible dialog is
  the §1 mask, still swallowing coordinate clicks. Report the family.
- **Mounting is not opening.** A standalone ModalOverlay (no
  DialogTrigger) renders closed until given an explicit `isOpen` —
  "clicked, no dialog in the DOM" is this family, not a lost event.
- **Keyboard dismissal is not dependable on RAC overlays.** filterDOMProps
  drops `onKeyDown*` props, and the library's Escape chain has broken
  releases — a failing Escape is an app bug family, not your `press_key`.
- **`isDismissable` plants a visually hidden cancel button** — a
  role+name query double-hits (snapshot or locator alike); the hidden
  twin is real accessibility, not a duplicate render.
