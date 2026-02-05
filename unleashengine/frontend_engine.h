#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

void *new_engine(void);

void free_engine(void *engine_ptr);

const char *take_state(void *engine_ptr, const char *json_ptr);

const uint8_t *resolve_all(void *engine_ptr,
                           const uint8_t *context_data,
                           const bool *include_all,
                           uintptr_t context_len,
                           uintptr_t *out_len);

const uint8_t *resolve(void *engine_ptr,
                       const char *toggle_name_ptr,
                       const uint8_t *context_data,
                       uintptr_t context_len,
                       uintptr_t *out_len);

bool is_enabled(void *engine_ptr,
                const char *toggle_name_ptr,
                const uint8_t *context_data,
                uintptr_t context_len);

void free_rust_buffer(uint8_t *ptr, uintptr_t len);

void free_response(char *response_ptr);
