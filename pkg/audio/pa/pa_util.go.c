#include <pulse/context.h>

extern void successCb(int success, void *userdata);
extern void moduleIDCb(uint idx, void *userdata);
extern void stateChanged (void* userdata);

void success_cb (pa_context *c, int success, void *userdata) {
	successCb(success, userdata);
};

void state_change_cb (pa_context *c, void *userdata) {
	stateChanged(userdata);
};

void new_module_cb (pa_context *c, uint32_t idx, void *userdata) {
	moduleIDCb(idx, userdata);
};