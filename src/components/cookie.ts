import { atom, computed } from 'nanostores';

const cookieLevelKey = 'cookie-level';
const cookieUpdateDate = 'cookie-update-date';

export const noLevel = -1;
export const rejectAllLevel = 0;
export const acceptOnlyRequiredLevel = 1;
export const acceptAllLevel = 2;

const level = atom<number>(loadLevel());
export const isBannerOpen = atom<boolean>(false);

export const isAllAccepted = computed(level, (level) => level === acceptAllLevel);
export const isAllRejected = computed(level, (level) => level === rejectAllLevel);
export const isOnlyRequiredAccepted = computed(level, (level) => level >= acceptOnlyRequiredLevel);
const isNoLevel = computed(level, (level) => level == noLevel);

isNoLevel.subscribe(openBannerIfNoLevel);

function openBannerIfNoLevel(isNoLevel: boolean) {
  if (!isNoLevel) {
    return;
  }
  isBannerOpen.set(true);
}

function loadLevel(): number {
  const lastUpdate = new Date(localStorage.getItem(cookieUpdateDate) || 0);

  if (new Date().getTime() - lastUpdate.getTime() > 1000 * 60 * 60 * 24 * 60) {
    return noLevel;
  }

  return Number(localStorage.getItem(cookieLevelKey) || noLevel);
}

function saveLevel(level: number) {
  localStorage.setItem(cookieUpdateDate, new Date().toISOString());
  localStorage.setItem(cookieLevelKey, level.toString());
}

export function rejectAll() {
  level.set(rejectAllLevel);
  isBannerOpen.set(false);
  saveLevel(rejectAllLevel);
}

export function acceptOnlyRequired() {
  level.set(acceptOnlyRequiredLevel);
  isBannerOpen.set(false);
  saveLevel(acceptOnlyRequiredLevel);
}

export function acceptAll() {
  level.set(acceptAllLevel);
  isBannerOpen.set(false);
  saveLevel(acceptAllLevel);
}
