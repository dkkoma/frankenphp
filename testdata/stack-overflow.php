<?php

function overflow(): void
{
    array_map('overflow', [0]);
}

overflow();
