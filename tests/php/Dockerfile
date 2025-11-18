FROM php:8.4

# Install system dependencies (git is needed for composer to download zips)
RUN apt-get update && apt-get install -y git zip unzip

# Install Composer
RUN php -r "copy('https://getcomposer.org/installer', 'composer-setup.php');" \
    && php composer-setup.php --install-dir=/usr/local/bin --filename=composer \
    && php -r "unlink('composer-setup.php');"

# Install Unleash Client AND the required implementations (Cache + HTTP)
# symfony/cache provides psr/simple-cache-implementation
# guzzlehttp/guzzle provides psr/http-client-implementation
RUN composer require unleash/client symfony/cache guzzlehttp/guzzle

COPY main.php main.php

CMD ["php", "main.php"]