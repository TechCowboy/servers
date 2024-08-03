
#define PROTOTYPES_ONLY
#include <smartkeys.h>


void sound_init(void)
{
    smartkeys_sound_init();
}

void sound_negative_beep(void)
{
    smartkeys_sound_play(SOUND_NEGATIVE_BUZZ);
}

void sound_chime(void)
{
    smartkeys_sound_play(SOUND_POSITIVE_CHIME);
}

void sound_mode_change(void)
{
    smartkeys_sound_play(SOUND_POSITIVE_CHIME);
}

void sound_confirm(void)
{
    smartkeys_sound_play(SOUND_LONG_BEEP);
}

