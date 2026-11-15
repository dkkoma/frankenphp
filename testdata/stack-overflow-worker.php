<?php

function overflow(): void
{
    array_map('overflow', [0]);
}

while (frankenphp_handle_request('overflow')) {
}
