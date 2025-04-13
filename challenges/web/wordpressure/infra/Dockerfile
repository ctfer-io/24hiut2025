FROM wordpress:latest

# Updating system.
RUN apt-get update \
    && apt-get install -y wget unzip \
    && rm -rf /var/lib/apt/lists/*

# Installing vulnerable plugin
RUN wget https://downloads.wordpress.org/plugin/custom-registration-form-builder-with-submission-manager.6.0.2.6.zip -P /var/www/html/wp-content/plugins/ \
    && unzip /var/www/html/wp-content/plugins/custom-registration-form-builder-with-submission-manager.6.0.2.6.zip -d /var/www/html/wp-content/plugins/ \
    && rm /var/www/html/wp-content/plugins/custom-registration-form-builder-with-submission-manager.6.0.2.6.zip \
    && mv /var/www/html/wp-content/plugins/custom-registration-form-builder-with-submission-manager /var/www/html/wp-content/plugins/registrationmagic

COPY hooked_class_rm_activator.php /var/www/html/wp-content/plugins/registrationmagic/includes/class_rm_activator.php
