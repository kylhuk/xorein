Deep Focus Chat

Product Overview

The Pitch: A unified communication hub engineered for zero-latency collaboration on any hardware. By stripping away computational bloat and focusing on high-contrast, deep-spectrum visuals, we create an environment where clarity is paramount and performance is absolute.

For: Enterprise teams and private communities who demand reliable, low-resource communication tools that don't sacrifice modern aesthetics.

Device: Desktop (Web/Electron)

Design Direction: Cyber-Industrial High-Fidelity. A refined dark theme utilizing deep indigo and charcoal substrates. Interactivity is signaled via sharp neon accents and distinct, pixel-perfect borders rather than expensive blur effects.

Inspired by: Linear’s precision, Discord’s layout, VS Code's efficiency.



Screens





Global Navigation: The omnipresent sidebar for server/workspace switching.



Channel Browser: The hierarchical list of text and voice channels within a workspace.



Chat Interface: The primary message stream with input and thread capabilities.



Voice HUD: The active call interface overlay for voice/video participants.



User Settings: Profile management, notifications, and accessibility controls.



Direct Messages: A focused list for 1:1 and private group conversations.



Key Flows

Flow: Joining a Voice Channel





User is on Channel Browser -> sees #general-voice with active users (green pulse).



User clicks #general-voice -> Voice HUD slides down from header; audio connects instantly.



Microphone status icon turns solid neon green; user avatar appears in the channel list.

Flow: Mentioning a Teammate





User is on Chat Interface -> types @ in input field.



Autocomplete Popover snaps into view immediately above cursor.



User selects name -> Text highlights in var(--color-accent); notification payload prepared.



Design System

Color Palette

Optimized for high contrast and low eye strain. No gradients on large surfaces to save rendering cost.





Primary: #6366F1 - Indigo (Primary Actions, active states)



Background: #0F1117 - Deepest Void (App background, Global Nav)



Surface: #1E2029 - Charcoal (Chat area, Sidebars)



Panel: #2A2D3A - Lighter Charcoal (Modals, Popovers, Hover states)



Text High: #E2E8F0 - Primary content (Slate 200)



Text Low: #94A3B8 - Meta data, timestamps (Slate 400)



Accent: #10B981 - Success/Online (Emerald Neon)



Alert: #F43F5E - Error/Mention/Busy (Rose Neon)



Border: #363B49 - distinct separation lines (No box-shadows used for depth)

Typography

Performance-focused variable fonts with distinct character.





Font Family: 'JetBrains Mono' (UI & Code) / 'Plus Jakarta Sans' (Body Copy)



H1 (Channel Titles): 'Plus Jakarta Sans', 700, 18px, tracking -0.01em



H2 (Section Headers): 'JetBrains Mono', 500, 12px, UPPERCASE, tracking 0.05em



Body (Messages): 'Plus Jakarta Sans', 400, 15px, line-height 1.5



UI (Labels/Buttons): 'JetBrains Mono', 500, 13px



Micro: 'JetBrains Mono', 400, 11px

Style notes: Hard edges (0px or 4px radius max). High contrast borders instead of drop shadows. "Glassy" look achieved via 95% opacity backgrounds + 1px solid borders, avoiding backdrop-filter: blur.

Design Tokens

:root {
  --color-primary: #6366F1;
  --color-bg-app: #0F1117;
  --color-bg-surface: #1E2029;
  --color-bg-panel: #2A2D3A;
  --color-text-high: #E2E8F0;
  --color-text-low: #94A3B8;
  --color-border: #363B49;
  --font-body: 'Plus Jakarta Sans', sans-serif;
  --font-mono: 'JetBrains Mono', monospace;
  --radius-sm: 2px;
  --radius-md: 4px;
  --spacing-unit: 4px;
}





Screen Specifications

Global Navigation (Sidebar Left)

Purpose: Root level navigation between different workspaces/servers.

Layout: Vertical column, 72px width, fixed left.

Key Elements:





Home Icon: Top, 48x48px container. Logo SVG.



Server Icons: 48x48px circles. Images with border-radius: 50%.



Active Pill: 4px wide, 40px high white bar to the left of the active server.



Separator: 2px solid #2A2D3A line separating Home from Servers.

States:





Unread: Small white dot (8px) on left of server icon.



Mentioned: Red badge with count on bottom-right of server icon.



Hover: Icon shifts from circle to border-radius: 16px (CSS transition: 0.2s ease).

Interactions:





Click Server: Switches main view context. Active Pill expands height.



Channel Browser (Sidebar Middle)

Purpose: Navigation within the selected workspace.

Layout: Vertical column, 240px width, right of Global Nav. Background #1E2029. Border right 1px #363B49.

Key Elements:





Header: 48px height. Server Name (H1). Chevron down icon.



Category Headers: Uppercase, mono font, #94A3B8, collapsed/expanded chevron.



Channel Item: Padding 8px 12px. # icon (text) or 🔊 icon (voice). Name in 'Plus Jakarta Sans'.

States:





Active Channel: Background #2A2D3A, Text #E2E8F0, 2px left border #6366F1.



Unread Channel: Text weight 700, Color #E2E8F0.



Muted: Text opacity 0.5.

Components:





User Control Panel: Bottom fixed. Avatar (32px), Name, Mic/Deafen toggles (icon buttons).



Chat Interface (Main View)

Purpose: The core conversation stream.

Layout: Flexible width. Background #1E2029.

Key Elements:





Top Bar: 48px height. Channel Name # general (H1). Topic text (truncate). Search Icon (right). Border bottom 1px #363B49.



Message List: Scrollable area. Messages grouped by user.





Avatar: 40px square, border-radius: 4px.



Header: Username (Color depending on role), Timestamp (Micro font).



Body: Markdown supported text. Code blocks in #0F1117.



Input Area: Bottom fixed.





Container: Margin 16px. Background #2A2D3A. Border 1px #363B49. Radius 4px.



Input: Textarea autosize. Placeholder "Message #channel".



Upload Button: + icon left.

States:





Loading: Skeleton bars pulsating opacity 0.5 to 1.0.



Empty Channel: "Welcome to #channel" illustration + "Start the conversation" CTA.

Interactions:





Hover Message: Message container background lightens slightly. Action toolbar (React, Reply) appears absolute positioned right.



Voice HUD (Overlay)

Purpose: Manage active voice participation without leaving chat context.

Layout: Collapsible panel within Channel Browser or Pop-out Grid.

Key Elements:





User Tile: Card, background #0F1117. Border 1px #363B49.



Speaking Indicator: Green border (#10B981) 2px glowing around User Tile.



Screen Share: "Live" badge (Red) on user tile.

States:





Muted: Red slash mic icon over avatar.



Connecting: Spinner on avatar.



User Settings (Modal)

Purpose: Account and app configuration.

Layout: Full screen overlay. Two columns: Sidebar (Tabs), Content (Form).

Key Elements:





Sidebar: "My Account", "Privacy", "Appearance". Active tab highlighted #2A2D3A.



Content Area:





Toggle Switch: 40px width. Track #0F1117, Thumb #6366F1.



Radio Cards: Selecting Theme. Options displayed as cards. Active card has #6366F1 border.



Close Button: ESC or top right 'X' icon circle.

Responsive:





Desktop: Modal 800x600px centered.



Mobile: Full screen push navigation.





Build Guide

Stack: HTML5 + Tailwind CSS v3 (JIT Mode enabled for performance).

Build Order:





Design System & Layout Shell: Define CSS variables for colors/fonts. Build the 3-column grid (Nav, Sidebar, Main). Crucial for establishing the "Deep Focus" atmosphere immediately.



Global Navigation: Implement server switching logic and tooltips.



Chat Interface: Build the message component. This is the most repeated element and determines rendering performance on low-end devices. Focus on efficient DOM recycling if lists get long.



Channel Browser: Implement collapsible categories and active states.



Voice HUD & Settings: Secondary layers.